package iostreams

import "testing"

func TestApplyJQ(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		expr    string
		want    string
		wantErr bool
	}{
		{
			name: "extract field",
			data: `{"result":[{"name":"测试","id":"123"},{"name":"hello","id":"456"}]}`,
			expr: `.result[].name`,
			want: "测试\nhello",
		},
		{
			name: "select with condition",
			data: `{"result":[{"name":"a","status":"ACTIVE"},{"name":"b","status":"CLOSED"}]}`,
			expr: `[.result[] | select(.status=="ACTIVE")] | length`,
			want: "1",
		},
		{
			name: "pipe to keys",
			data: `{"foo":1,"bar":2}`,
			expr: `keys`,
			want: `["bar","foo"]`,
		},
		{
			name: "chinese characters preserved",
			data: `{"entityName":"测试一下设备"}`,
			expr: `.entityName`,
			want: "测试一下设备",
		},
		{
			name: "null result",
			data: `{"a":1}`,
			expr: `.b`,
			want: "null",
		},
		{
			name:    "invalid expression",
			data:    `{}`,
			expr:    `.[invalid`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ApplyJQ([]byte(tt.data), tt.expr)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
