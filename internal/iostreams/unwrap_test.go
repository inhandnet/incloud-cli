package iostreams

import "testing"

func TestUnwrapResult(t *testing.T) {
	tests := []struct {
		name string
		data string
		want string
	}{
		{
			name: "single key result object — unwrap",
			data: `{"result":{"name":"test","id":"123"}}`,
			want: `{"name":"test","id":"123"}`,
		},
		{
			name: "single key result array — unwrap",
			data: `{"result":[{"name":"a"},{"name":"b"}]}`,
			want: `[{"name":"a"},{"name":"b"}]`,
		},
		{
			name: "multi key envelope — keep as-is",
			data: `{"result":[{"name":"a"}],"total":1,"page":0,"limit":20}`,
			want: `{"result":[{"name":"a"}],"total":1,"page":0,"limit":20}`,
		},
		{
			name: "no result key — keep as-is",
			data: `{"count":5,"version":{"V2.0":3}}`,
			want: `{"count":5,"version":{"V2.0":3}}`,
		},
		{
			name: "single key but not result — keep as-is",
			data: `{"data":{"name":"test"}}`,
			want: `{"data":{"name":"test"}}`,
		},
		{
			name: "not an object — keep as-is",
			data: `[1,2,3]`,
			want: `[1,2,3]`,
		},
		{
			name: "invalid json — keep as-is",
			data: `not json`,
			want: `not json`,
		},
		{
			name: "result is null — unwrap",
			data: `{"result":null}`,
			want: `null`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := string(unwrapResult([]byte(tt.data)))
			if got != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}
