package gplay

// AuthBundle is the response shape Aurora Store clients expect.
// Field names intentionally match the original NodeJS dispenser.
type AuthBundle struct {
	AASToken                      string             `json:"aasToken"`
	AC2DMToken                    string             `json:"ac2dmToken"`
	AndroidCheckInToken           string             `json:"androidCheckInToken"`
	AuthToken                     string             `json:"authToken"`
	DeviceCheckInConsistencyToken string             `json:"deviceCheckInConsistencyToken"`
	DeviceConfigToken             string             `json:"deviceConfigToken"`
	DeviceInfoProvider            DeviceInfoProvider `json:"deviceInfoProvider"`
	DfeCookie                     string             `json:"dfeCookie"`
	Email                         string             `json:"email"`
	ExperimentsConfigToken        string             `json:"experimentsConfigToken"`
	GsfID                         string             `json:"gsfId"`
	IsAnonymous                   bool               `json:"isAnonymous"`
	Locale                        string             `json:"locale"`
	TokenDispenserURL             string             `json:"tokenDispenserUrl"`
	UserProfile                   UserProfile        `json:"userProfile"`
}

type AnonymousAuthBundle struct {
	Email string `json:"email"`
	Auth  string `json:"auth"`
}

type DeviceInfoProvider struct {
	AuthUserAgentString string       `json:"authUserAgentString"`
	LocaleString        string       `json:"localeString"`
	MccMnc              string       `json:"mccMnc"`
	PlayServicesVersion int          `json:"playServicesVersion"`
	Properties          DeviceConfig `json:"properties"`
	UserAgentString     string       `json:"userAgentString"`
	SDKVersion          int          `json:"sdkVersion"`
}

type UserProfile struct {
	Name    string  `json:"name"`
	Email   string  `json:"email"`
	Artwork Artwork `json:"artwork"`
}

type Artwork struct {
	URL    string `json:"url"`
	Type   int    `json:"type"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

func anonymousProfile() UserProfile {
	return UserProfile{
		Name:  "Anonymous",
		Email: "anonymous@gmail.com",
		Artwork: Artwork{
			URL:    "https://lh3.googleusercontent.com/a/default-user",
			Type:   4,
			Width:  96,
			Height: 96,
		},
	}
}

// Account is the minimal credential pair needed to mint a token.
type Account struct {
	Email    string
	AASToken string
}
