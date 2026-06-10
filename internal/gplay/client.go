package gplay

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"gplaydl-dispenser/internal/pb"

	"google.golang.org/protobuf/proto"
)

const (
	baseURL               = "https://android.clients.google.com"
	checkinURL            = baseURL + "/checkin"
	authURL               = baseURL + "/auth"
	uploadDeviceConfigURL = baseURL + "/fdfe/uploadDeviceConfig"
	tocURL                = baseURL + "/fdfe/toc"

	clientSignature = "38918a453d07199354f8b19af05ec6562ced5788"
)

// Client performs the Google Play handshake. Safe for concurrent use; a
// semaphore caps simultaneous handshakes so a burst can't exhaust sockets.
type Client struct {
	http              *http.Client
	sem               chan struct{}
	tokenDispenserURL string
}

func NewClient(maxConcurrent int, mintTimeout time.Duration, tokenDispenserURL string) *Client {
	if maxConcurrent < 1 {
		maxConcurrent = 1
	}
	transport := &http.Transport{
		MaxIdleConns:        maxConcurrent * 2,
		MaxIdleConnsPerHost: maxConcurrent,
		MaxConnsPerHost:     maxConcurrent * 2,
		IdleConnTimeout:     90 * time.Second,
		ForceAttemptHTTP2:   true,
	}
	return &Client{
		http: &http.Client{
			Transport: transport,
			Timeout:   mintTimeout,
		},
		sem:               make(chan struct{}, maxConcurrent),
		tokenDispenserURL: tokenDispenserURL,
	}
}

type authPayload struct {
	gsfID                  string
	userAgent              string
	deviceConsistencyToken string
	deviceConfigToken      string
	dfeCookie              string
}

// Mint runs the full 4-step handshake and returns an AuthBundle.
func (c *Client) Mint(ctx context.Context, account Account, dc DeviceConfig, locale string) (*AuthBundle, error) {
	select {
	case c.sem <- struct{}{}:
		defer func() { <-c.sem }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	userAgent := BuildUserAgent(dc)

	checkinResp, err := c.checkIn(ctx, dc, userAgent)
	if err != nil {
		return nil, fmt.Errorf("checkin: %w", err)
	}

	gsfID := strconv.FormatUint(checkinResp.GetAndroidId(), 16)
	payload := authPayload{
		gsfID:                  gsfID,
		userAgent:              userAgent,
		deviceConsistencyToken: checkinResp.GetDeviceCheckinConsistencyToken(),
	}

	deviceConfigToken, err := c.uploadDeviceConfig(ctx, dc, payload, locale)
	if err != nil {
		return nil, fmt.Errorf("uploadDeviceConfig: %w", err)
	}
	payload.deviceConfigToken = deviceConfigToken

	authToken, err := c.exchangeAASToken(ctx, account, dc, gsfID, locale)
	if err != nil {
		return nil, fmt.Errorf("auth: %w", err)
	}

	dfeCookie, err := c.acceptTOC(ctx, payload, authToken, locale)
	if err != nil {
		return nil, fmt.Errorf("toc: %w", err)
	}

	return &AuthBundle{
		AASToken:                      "REDACTED",
		AuthToken:                     authToken,
		DeviceCheckInConsistencyToken: payload.deviceConsistencyToken,
		DeviceConfigToken:             deviceConfigToken,
		DfeCookie:                     dfeCookie,
		GsfID:                         gsfID,
		IsAnonymous:                   false,
		Locale:                        locale,
		TokenDispenserURL:             c.tokenDispenserURL,
		Email:                         account.Email,
		DeviceInfoProvider: DeviceInfoProvider{
			AuthUserAgentString: authUserAgent(dc),
			LocaleString:        dc.Get("locale"),
			MccMnc:              "310260",
			PlayServicesVersion: dc.Int("GSF.version"),
			UserAgentString:     userAgent,
			SDKVersion:          dc.Int("Build.VERSION.SDK_INT"),
			Properties:          dc,
		},
		UserProfile: anonymousProfile(),
	}, nil
}

func (c *Client) checkIn(ctx context.Context, dc DeviceConfig, userAgent string) (*pb.AndroidCheckinResponse, error) {
	body, err := proto.Marshal(BuildCheckinRequest(dc))
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, checkinURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("app", "com.google.android.gms")
	req.Header.Set("Content-Type", "application/x-protobuffer")
	req.Header.Set("User-Agent", userAgent)

	data, err := c.do(req)
	if err != nil {
		return nil, err
	}

	resp := &pb.AndroidCheckinResponse{}
	if err := proto.Unmarshal(data, resp); err != nil {
		return nil, err
	}
	if resp.GetAndroidId() == 0 {
		return nil, fmt.Errorf("checkin returned no androidId")
	}
	return resp, nil
}

func (c *Client) uploadDeviceConfig(ctx context.Context, dc DeviceConfig, payload authPayload, locale string) (string, error) {
	body, err := proto.Marshal(&pb.UploadDeviceConfigRequest{
		DeviceConfiguration: BuildDeviceConfigProto(dc),
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uploadDeviceConfigURL, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	setDFEHeaders(req, payload, locale, "")
	req.Header.Set("Content-Type", "application/x-protobuf")

	data, err := c.do(req)
	if err != nil {
		return "", err
	}

	wrapper := &pb.ResponseWrapper{}
	if err := proto.Unmarshal(data, wrapper); err != nil {
		return "", err
	}
	return wrapper.GetPayload().GetUploadDeviceConfigResponse().GetUploadDeviceConfigToken(), nil
}

// exchangeAASToken trades a long-lived AAS token for a Play Store OAuth token.
func (c *Client) exchangeAASToken(ctx context.Context, account Account, dc DeviceConfig, gsfID, locale string) (string, error) {
	params := url.Values{
		"app":                          {"com.android.vending"},
		"oauth2_foreground":            {"1"},
		"Email":                        {account.Email},
		"token_request_options":        {"CAA4AVAB"},
		"client_sig":                   {clientSignature},
		"Token":                        {account.AASToken},
		"google_play_services_version": {dc.Get("GSF.version")},
		"check_email":                  {"1"},
		"system_partition":             {"1"},
		"sdk_version":                  {dc.Get("Build.VERSION.SDK_INT")},
		"callerPkg":                    {"com.google.android.gms"},
		"device_country":               {"IN"},
		"lang":                         {locale},
		"androidId":                    {gsfID},
		"callerSig":                    {clientSignature},
		"service":                      {"oauth2:https://www.googleapis.com/auth/googleplay"},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, authURL+"?"+params.Encode(), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("app", "com.google.android.gms")
	req.Header.Set("device", gsfID)
	req.Header.Set("User-Agent", authUserAgent(dc))

	data, err := c.do(req)
	if err != nil {
		return "", err
	}

	kv := parseKeyValues(string(data))
	if auth := kv["Auth"]; auth != "" {
		return auth, nil
	}
	if e := kv["Error"]; e != "" {
		return "", fmt.Errorf("google rejected credentials: %s", e)
	}
	return "", fmt.Errorf("no Auth token in response")
}

func (c *Client) acceptTOC(ctx context.Context, payload authPayload, bearerToken, locale string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, tocURL, nil)
	if err != nil {
		return "", err
	}
	setDFEHeaders(req, payload, locale, bearerToken)

	data, err := c.do(req)
	if err != nil {
		return "", err
	}

	wrapper := &pb.ResponseWrapper{}
	if err := proto.Unmarshal(data, wrapper); err != nil {
		return "", err
	}
	return wrapper.GetPayload().GetTocResponse().GetCookie(), nil
}

func (c *Client) do(req *http.Request) ([]byte, error) {
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		msg := strings.TrimSpace(string(data))
		if len(msg) > 200 {
			msg = msg[:200]
		}
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, msg)
	}
	return data, nil
}

func authUserAgent(dc DeviceConfig) string {
	return fmt.Sprintf("GoogleAuth/1.4 (%s %s)", dc.Get("Build.DEVICE"), dc.Get("Build.ID"))
}

func setDFEHeaders(req *http.Request, payload authPayload, locale, bearerToken string) {
	h := req.Header
	if bearerToken != "" {
		h.Set("Authorization", "Bearer "+bearerToken)
	}
	h.Set("X-DFE-Encoded-Targets", "")
	h.Set("User-Agent", payload.userAgent)
	h.Set("X-DFE-Cookie", payload.dfeCookie)
	h.Set("X-DFE-Content-Filters", "")
	h.Set("X-DFE-Device-Checkin-Consistency-Token", payload.deviceConsistencyToken)
	h.Set("X-DFE-Device-Config-Token", payload.deviceConfigToken)
	h.Set("X-DFE-MCCMNC", "21601")
	h.Set("X-DFE-Client-Id", "am-android-google")
	h.Set("X-DFE-UserLanguages", locale)
	h.Set("X-DFE-Phenotype", "")
	h.Set("X-DFE-Device-Id", payload.gsfID)
	h.Set("X-DFE-Network-Type", "4")
	h.Set("Accept-Language", locale)
	h.Set("X-DFE-Request-Params", "timeoutMs=4000")
	h.Set("X-DFE-Enabled-Experiments", "cl:billing.select_add_instrument_by_default")
	h.Set("X-DFE-Unsupported-Experiments", "nocache:billing.use_charging_poller,market_emails,buyer_currency,prod_baseline,checkin.set_asset_paid_app_field,shekel_test,content_ratings,buyer_currency_in_app,nocache:encrypted_apk,recent_changes")
}

func parseKeyValues(body string) map[string]string {
	out := map[string]string{}
	for _, line := range strings.Split(body, "\n") {
		if k, v, ok := strings.Cut(line, "="); ok {
			out[strings.TrimSpace(k)] = strings.TrimSpace(v)
		}
	}
	return out
}
