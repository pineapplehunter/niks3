package main

import "testing"

func TestSelectOutputMode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		value    string
		terminal bool
		want     outputMode
		wantErr  bool
	}{
		{name: "auto terminal", value: "auto", terminal: true, want: outputProgress},
		{name: "auto redirected", value: "auto", terminal: false, want: outputLog},
		{name: "log", value: "log", terminal: true, want: outputLog},
		{name: "progress", value: "progress", terminal: false, want: outputProgress},
		{name: "invalid", value: "quiet", wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got, err := selectOutputMode(test.value, test.terminal)
			if (err != nil) != test.wantErr {
				t.Fatalf("selectOutputMode() error = %v, wantErr %v", err, test.wantErr)
			}

			if got != test.want {
				t.Errorf("selectOutputMode() = %q, want %q", got, test.want)
			}
		})
	}
}
