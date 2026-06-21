package publication

import (
	"testing"

	"github.com/realmkit/rk-backend/module/themes/domain"
)

// benchmarkActivationSettings stores publication settings benchmark output.
var benchmarkActivationSettings []byte

// BenchmarkActivationSettings measures required settings validation and normalization before activation.
func BenchmarkActivationSettings(b *testing.B) {
	version := domain.ThemeVersion{
		SettingsSchemaJSON: []byte(`{"required":["brand","accent","locale"]}`),
		SettingsDataJSON:   []byte(`{"brand":"RealmKit","accent":"#33cc99","locale":"en"}`),
	}
	override := []byte(`{"brand":"RealmKit","accent":"#33cc99","locale":"en","layout":"wide"}`)

	b.ReportAllocs()
	for index := 0; index < b.N; index++ {
		settings, err := activationSettings(version, override)
		if err != nil {
			b.Fatalf("activationSettings() error = %v", err)
		}
		benchmarkActivationSettings = settings
	}
}
