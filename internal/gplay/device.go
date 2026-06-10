package gplay

import (
	"fmt"
	"strings"
	"time"

	"gplaydl-dispenser/internal/pb"

	"google.golang.org/protobuf/proto"
)

// BuildUserAgent assembles the Finsky (Play Store) user agent string.
func BuildUserAgent(dc DeviceConfig) string {
	platforms := strings.Join(dc.List("Platforms"), ";")

	pairs := []string{
		"api=3",
		fmt.Sprintf("versionCode=%s", dc.Get("Vending.version")),
		fmt.Sprintf("sdk=%s", dc.Get("Build.VERSION.SDK_INT")),
		fmt.Sprintf("device=%s", dc.Get("Build.DEVICE")),
		fmt.Sprintf("hardware=%s", dc.Get("Build.HARDWARE")),
		fmt.Sprintf("product=%s", dc.Get("Build.PRODUCT")),
		fmt.Sprintf("platformVersionRelease=%s", dc.Get("Build.VERSION.RELEASE")),
		fmt.Sprintf("model=%s", dc.Get("Build.MODEL")),
		fmt.Sprintf("buildId=%s", dc.Get("Build.ID")),
		"isWideScreen=0",
		fmt.Sprintf("supportedAbis=%s", platforms),
	}

	return fmt.Sprintf("Android-Finsky/%s (%s)", dc.Get("Vending.versionString"), strings.Join(pairs, ","))
}

// BuildDeviceConfigProto mirrors the device capability payload Play expects.
func BuildDeviceConfigProto(dc DeviceConfig) *pb.DeviceConfigurationProto {
	features := dc.List("Features")
	deviceFeatures := make([]*pb.DeviceFeature, 0, len(features))
	for _, f := range features {
		deviceFeatures = append(deviceFeatures, &pb.DeviceFeature{
			Name:  proto.String(f),
			Value: proto.Int32(0),
		})
	}

	return &pb.DeviceConfigurationProto{
		TouchScreen:            proto.Int32(int32(dc.Int("TouchScreen"))),
		Keyboard:               proto.Int32(int32(dc.Int("Keyboard"))),
		Navigation:             proto.Int32(int32(dc.Int("Navigation"))),
		ScreenLayout:           proto.Int32(int32(dc.Int("ScreenLayout"))),
		HasHardKeyboard:        proto.Bool(dc.Bool("HasHardKeyboard")),
		HasFiveWayNavigation:   proto.Bool(dc.Bool("HasFiveWayNavigation")),
		LowRamDevice:           proto.Int32(boolToInt32(dc.Bool("LowRamDevice"))),
		MaxNumOf_CPUCores:      proto.Int32(int32(dc.Int("MaxNumOfCPUCores"))),
		TotalMemoryBytes:       proto.Int64(dc.Int64("TotalMemoryBytes")),
		GlEsVersion:            proto.Int32(int32(dc.Int("GL.Version"))),
		GlExtension:            dc.List("GL.Extensions"),
		SystemSharedLibrary:    dc.List("SharedLibraries"),
		SystemAvailableFeature: features,
		NativePlatform:         dc.List("Platforms"),
		ScreenDensity:          proto.Int32(int32(dc.Int("Screen.Density"))),
		ScreenWidth:            proto.Int32(int32(dc.Int("Screen.Width"))),
		ScreenHeight:           proto.Int32(int32(dc.Int("Screen.Height"))),
		SystemSupportedLocale:  dc.List("Locales"),
		DeviceClass:            proto.Int32(0),
		DeviceFeature:          deviceFeatures,
	}
}

// BuildCheckinRequest assembles the initial device check-in.
func BuildCheckinRequest(dc DeviceConfig) *pb.AndroidCheckinRequest {
	build := &pb.AndroidBuildProto{
		Id:             proto.String(dc.Get("Build.FINGERPRINT")),
		Product:        proto.String(dc.Get("Build.HARDWARE")),
		Carrier:        proto.String(dc.Get("Build.BRAND")),
		Radio:          proto.String(dc.Get("Build.RADIO")),
		Bootloader:     proto.String(dc.Get("Build.BOOTLOADER")),
		Device:         proto.String(dc.Get("Build.DEVICE")),
		SdkVersion:     proto.Int32(int32(dc.Int("Build.VERSION.SDK_INT"))),
		Model:          proto.String(dc.Get("Build.MODEL")),
		Manufacturer:   proto.String(dc.Get("Build.MANUFACTURER")),
		BuildProduct:   proto.String(dc.Get("Build.PRODUCT")),
		Client:         proto.String(dc.Get("Client")),
		OtaInstalled:   proto.Bool(dc.Bool("OtaInstalled")),
		Timestamp:      proto.Int64(time.Now().UnixMilli()),
		GoogleServices: proto.Int32(int32(dc.Int("GSF.version"))),
	}

	checkin := &pb.AndroidCheckinProto{
		Build:           build,
		LastCheckinMsec: proto.Int64(0),
		CellOperator:    proto.String(dc.Get("CellOperator")),
		SimOperator:     proto.String(dc.Get("SimOperator")),
		Roaming:         proto.String(dc.Get("Roaming")),
		UserNumber:      proto.Int32(0),
	}

	return &pb.AndroidCheckinRequest{
		Id:                  proto.Int64(0),
		Checkin:             checkin,
		Locale:              proto.String("en"),
		TimeZone:            proto.String(dc.Get("TimeZone")),
		Version:             proto.Int32(3),
		DeviceConfiguration: BuildDeviceConfigProto(dc),
		Fragment:            proto.Int32(0),
	}
}

func boolToInt32(b bool) int32 {
	if b {
		return 1
	}
	return 0
}
