package drain

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type TestCase struct {
	name string
	line string
	want map[string][]string
}

const (
	typePunctuation = "punctuation"
	typeSplitting   = "splitting"
)

var testCases = []TestCase{
	{
		name: "Test with equals sign",
		line: "key1=value1 key2=value2",
		want: map[string][]string{
			typePunctuation: {"key1", "=", "value1", "key2", "=", "value2"},
			typeSplitting:   {"key1=", "value1", "key2=", "value2"},
		},
	},
	{
		name: "Test with colon",
		line: "key1:value1 key2:value2",
		want: map[string][]string{
			typePunctuation: {"key1:value1", "key2:value2"},
			typeSplitting:   {"key1:", "value1", "key2:", "value2"},
		},
	},
	{
		name: "Test with mixed delimiters, more = than :",
		line: "key1=value1 key2:value2 key3=value3",
		want: map[string][]string{
			typePunctuation: {"key1", "=", "value1", "key2:value2", "key3", "=", "value3"},
			typeSplitting:   {"key1=", "value1", "key2:value2", "key3=", "value3"},
		},
	},
	{
		name: "Test with mixed delimiters, more : than =",
		line: "key1:value1 key2:value2 key3=value3",
		want: map[string][]string{
			typePunctuation: {"key1:value1", "key2:value2", "key3", "=", "value3"},
			typeSplitting:   {"key1:", "value1", "key2:", "value2", "key3=value3"},
		},
	},
	{
		name: "Dense json",
		line: `{"key1":"value1","key2":"value2","key3":"value3"}`,
		want: map[string][]string{
			typePunctuation: {`{`, `"`, `key1`, `"`, `:`, `"`, `value1`, `"`, `,`, `"`, `key2`, `"`, `:`, `"`, `value2`, `"`, `,`, `"`, `key3`, `"`, `:`, `"`, `value3`, `"`, `}`},
			typeSplitting:   {`{"key1":`, `"value1","key2":`, `"value2","key3":`, `"value3"}`},
		},
	},
	{
		name: "json with spaces",
		line: `{"key1":"value1", "key2":"value2", "key3":"value3"}`,
		want: map[string][]string{
			typePunctuation: {`{`, `"`, `key1`, `"`, `:`, `"`, `value1`, `"`, `,`, `"`, `key2`, `"`, `:`, `"`, `value2`, `"`, `,`, `"`, `key3`, `"`, `:`, `"`, `value3`, `"`, `}`},
			typeSplitting:   {`{"key1":`, `"value1",`, `"key2":`, `"value2",`, `"key3":`, `"value3"}`},
		},
	},
	{
		name: "logfmt multiword values",
		line: `key1=value1 key2=value2 msg="this is a message"`,
		want: map[string][]string{
			typePunctuation: {"key1", "=", "value1", "key2", "=", "value2", "msg", "=", `"`, `this`, "is", "a", `message`, `"`},
			typeSplitting:   {"key1=", "value1", "key2=", "value2", "msg=", `"this`, "is", "a", `message"`},
		},
	},
	{
		name: "longer line",
		line: "09:17:38.033366 ▶ INFO  route ops sending to dest https://graphite-cortex-ops-blocks-us-east4.grafana.net/graphite/metrics: service_is_carbon-relay-ng.instance_is_carbon-relay-ng-c665b7b-j2trk.mtype_is_counter.dest_is_https_graphite-cortex-ops-blocks-us-east4_grafana_netgraphitemetrics.unit_is_Metric.action_is_drop.reason_is_queue_full 0 1717060658",
		want: map[string][]string{
			typePunctuation: {`09:17:38.033366`, `▶`, `INFO`, `route`, `ops`, `sending`, `to`, `dest`, `https://graphite-cortex-ops-blocks-us-east4.grafana.net/graphite/metrics:`, `service_is_carbon-relay-ng.instance_is_carbon-relay-ng-c665b7b-j2trk.mtype_is_counter.dest_is_https_graphite-cortex-ops-blocks-us-east4_grafana_netgraphitemetrics.unit_is_Metric.action_is_drop.reason_is_queue_full`, `0`, `1717060658`},
			typeSplitting:   {`09:`, `17:`, `38.033366`, `▶`, `INFO`, ``, `route`, `ops`, `sending`, `to`, `dest`, `https:`, `//graphite-cortex-ops-blocks-us-east4.grafana.net/graphite/metrics:`, ``, `service_is_carbon-relay-ng.instance_is_carbon-relay-ng-c665b7b-j2trk.mtype_is_counter.dest_is_https_graphite-cortex-ops-blocks-us-east4_grafana_netgraphitemetrics.unit_is_Metric.action_is_drop.reason_is_queue_full`, `0`, `1717060658`},
		},
	},
	{
		name: "Consecutive splits points: equals followed by space",
		line: `ts=2024-05-30T12:50:36.648377186Z caller=scheduler_processor.go:143 level=warn msg="error contacting scheduler" err="rpc error: code = Unavailable desc = connection error: desc = \"error reading server preface: EOF\"" addr=10.0.151.101:9095`,
		want: map[string][]string{
			typePunctuation: {`ts`, `=`, `2024-05-30T12:50:36.648377186Z`, `caller`, `=`, `scheduler_processor.go:143`, `level`, `=`, `warn`, `msg`, `=`, `"`, `error`, `contacting`, `scheduler`, `"`, `err`, `=`, `"`, `rpc`, `error:`, `code`, `=`, `Unavailable`, `desc`, `=`, `connection`, `error:`, `desc`, `=`, `\`, `"`, `error`, `reading`, `server`, `preface:`, `EOF`, `\`, `"`, `"`, `addr`, `=`, `10.0.151.101:9095`},
			typeSplitting:   {"ts=", "2024-05-30T12:50:36.648377186Z", "caller=", "scheduler_processor.go:143", "level=", "warn", "msg=", "\"error", "contacting", "scheduler\"", "err=", "\"rpc", "error:", "code", "=", ``, "Unavailable", "desc", "=", ``, "connection", "error:", "desc", "=", ``, `\"error`, "reading", "server", "preface:", `EOF\""`, "addr=", "10.0.151.101:9095"},
		},
	},
	{
		name: "Exactly 128 tokens are not combined",
		line: strings.Repeat(`A `, 126) + "127 128",
		want: map[string][]string{
			typePunctuation: append(strings.Split(strings.Repeat(`A `, 126), " ")[:126], "127", "128"),
			typeSplitting:   append(strings.Split(strings.Repeat(`A `, 126), " ")[:126], "127", "128"),
		},
	},
	{
		name: "More than 128 tokens combined suffix into one token",
		line: strings.Repeat(`A `, 126) + "127 128 129",
		want: map[string][]string{
			typePunctuation: append(strings.Split(strings.Repeat(`A `, 126), " ")[:126], "127", "128 129"),
			typeSplitting:   append(strings.Split(strings.Repeat(`A `, 126), " ")[:126], "127", "128", "129"),
		},
	},
	{
		name: "Only punctation",
		line: `!@£$%^&*()`,
		want: map[string][]string{
			typePunctuation: {`!`, `@`, `£$`, `%`, `^`, `&`, `*`, `(`, `)`},
			typeSplitting:   {`!@£$%^&*()`},
		},
	},
}

func TestTokenizer_Tokenize(t *testing.T) {
	tests := []struct {
		name      string
		tokenizer LineTokenizer
	}{
		{
			name:      typePunctuation,
			tokenizer: newPunctuationTokenizer(),
		},
		{
			name:      typeSplitting,
			tokenizer: splittingTokenizer{},
		},
	}

	for _, tt := range tests {
		for _, tc := range testCases {
			t.Run(tt.name+":"+tc.name, func(t *testing.T) {
				got, _ := tt.tokenizer.Tokenize(tc.line, nil, nil)
				require.Equal(t, tc.want[tt.name], got)
			})
		}
	}
}

func TestTokenizer_TokenizeAndJoin(t *testing.T) {
	tests := []struct {
		name      string
		tokenizer LineTokenizer
	}{
		{
			name:      typePunctuation,
			tokenizer: newPunctuationTokenizer(),
		},
		{
			name:      typeSplitting,
			tokenizer: splittingTokenizer{},
		},
	}

	for _, tt := range tests {
		for _, tc := range testCases {
			t.Run(tt.name+":"+tc.name, func(t *testing.T) {
				got := tt.tokenizer.Join(tt.tokenizer.Tokenize(tc.line, nil, nil))
				require.Equal(t, tc.line, got)
			})
		}
	}
}

func BenchmarkSplittingTokenizer(b *testing.B) {
	tokenizer := newPunctuationTokenizer()

	for _, tt := range testCases {
		tc := tt
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				tokenizer.Tokenize(tc.line, nil, nil)
			}
		})
	}
}

func TestLogFmtTokenizer(t *testing.T) {
	param := DefaultConfig().ParamString
	tests := []struct {
		name string
		line string
		want []string
	}{
		{
			line: `foo=bar baz="this is a message"`,
			want: []string{"foo", "bar", "baz", "this is a message"},
		},
		{
			line: `foo baz="this is a message"`,
			want: []string{"foo", "", "baz", "this is a message"},
		},
		{
			line: `foo= baz="this is a message"`,
			want: []string{"foo", "", "baz", "this is a message"},
		},
		{
			line: `foo baz`,
			want: []string{"foo", "", "baz", ""},
		},
		{
			line: `ts=2024-05-30T12:50:36.648377186Z caller=scheduler_processor.go:143 level=warn msg="error contacting scheduler" err="rpc error: code = Unavailable desc = connection error: desc = \"error reading server preface: EOF\"" addr=10.0.151.101:9095`,
			want: []string{"ts", param, "caller", "scheduler_processor.go:143", "level", "warn", "msg", "error contacting scheduler", "err", "rpc error: code = Unavailable desc = connection error: desc = \"error reading server preface: EOF\"", "addr", "10.0.151.101:9095"},
		},
	}

	tokenizer := newLogfmtTokenizer(param)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := tokenizer.Tokenize(tt.line, nil, nil)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestLogFmtTokenizerJoin(t *testing.T) {
	tests := []struct {
		tokens []string
		want   string
	}{
		{
			want:   ``,
			tokens: []string{},
		},
		{
			want:   `foo=bar baz="this is a message"`,
			tokens: []string{"foo", "bar", "baz", "this is a message"},
		},
		{
			want:   `foo= baz="this is a message"`,
			tokens: []string{"foo", "", "baz", "this is a message"},
		},
		{
			want:   `foo= baz="this is a message"`,
			tokens: []string{"foo", "", "baz", "this is a message"},
		},
		{
			want:   `foo= baz=`,
			tokens: []string{"foo", "", "baz", ""},
		},
		{
			want:   `foo=`,
			tokens: []string{"foo"},
		},
		{
			want:   `foo= bar=`,
			tokens: []string{"foo", "", "bar"},
		},
		{
			want:   `ts=2024-05-30T12:50:36.648377186Z caller=scheduler_processor.go:143 level=warn msg="error contacting scheduler" err="rpc error: code = Unavailable desc = connection error: desc = \"error reading server preface: EOF\"" addr=10.0.151.101:9095`,
			tokens: []string{"ts", "2024-05-30T12:50:36.648377186Z", "caller", "scheduler_processor.go:143", "level", "warn", "msg", "error contacting scheduler", "err", "rpc error: code = Unavailable desc = connection error: desc = \"error reading server preface: EOF\"", "addr", "10.0.151.101:9095"},
		},
	}

	tokenizer := newLogfmtTokenizer("")

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := tokenizer.Join(tt.tokens, nil)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestJsonTokenizer(t *testing.T) {
	param := DefaultConfig().ParamString
	tests := []struct {
		name    string
		line    string
		pattern string
		want    []string
	}{
		{
			line:    `{"level":30,"time":1719998371869,"pid":17,"hostname":"otel-demo-ops-paymentservice-7c759bf575-55t4p","trace_id":"1425c6df5a4321cf6a7de254de5b8204","span_id":"2ac7a3fc800b80d4","trace_flags":"01","transactionId":"e501032b-3215-4e43-b1db-f4886a906fc5","cardType":"visa","lastFourDigits":"5647","amount":{"units":{"low":656,"high":0,"unsigned":false},"nanos":549999996,"currencyCode":"USD"},"msg":"Transaction complete."}`,
			want:    []string{"Transaction", "complete."},
			pattern: "<_>Transaction complete.<_>",
		},
		{
			line:    `{"event":{"actor":{"alternateId":"foo@grafana.com","displayName":"Foo bar","id":"dq23","type":"User"},"authenticationContext":{"authenticationStep":0,"externalSessionId":"123d"},"client":{"device":"Computer","geographicalContext":{"city":"Berlin","country":"DE","state":"Land Berlin"},"ipAddress":"0.0.0.0","userAgent":{"browser":"CHROME","os":"Mac OS X","rawUserAgent":"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36"},"zone":"null"},"debugContext":{"debugData":{"authMethodFirstEnrollment":"123","authMethodFirstType":"foo","authMethodFirstVerificationTime":"2024-07-02T11:28:03.219Z","authMethodSecondEnrollment":"var","authMethodSecondType":"ddd","authMethodSecondVerificationTime":"2024-07-03T06:59:09.151Z","authnRequestId":"1","dtHash":"1","logOnlySecurityData":"{\"risk\":{\"level\":\"LOW\"},\"behaviors\":{\"New Geo-Location\":\"NEGATIVE\",\"New Device\":\"NEGATIVE\",\"New IP\":\"NEGATIVE\",\"New State\":\"NEGATIVE\",\"New Country\":\"NEGATIVE\",\"Velocity\":\"NEGATIVE\",\"New City\":\"NEGATIVE\"}}","requestId":"1","threatSuspected":"false","url":"/foo?"}},"displayMessage":"Evaluation of sign-on policy","eventType":"policy.evaluate_sign_on","legacyEventType":"app.oauth2.token.grant.refresh_token_success","outcome":{"reason":"Sign-on policy evaluation resulted in AUTHENTICATED","result":"ALLOW"},"published":"2024-07-03T09:19:59.973Z","request":{"ipChain":[{"geographicalContext":{"city":"Berlin","country":"Germany","geolocation":{"lat":52.5363,"lon":13.4169},"postalCode":"10435","state":"Land Berlin"},"ip":"95.90.234.241","version":"V4"}]},"securityContext":{"asNumber":3209,"asOrg":"kabel deutschland breitband customer 19","domain":"kabel-deutschland.de","isProxy":false,"isp":"vodafone gmbh"},"severity":"INFO","target":[{"alternateId":"Salesforce.com","detailEntry":{"signOnModeEvaluationResult":"AUTHENTICATED","signOnModeType":"SAML_2_0"},"displayName":"Salesforce.com","id":"0oa5sfmj3hz0mTgoW357","type":"AppInstance"},{"alternateId":"unknown","detailEntry":{"policyRuleFactorMode":"2FA"},"displayName":"Catch-all Rule","id":"1","type":"Rule"}],"transaction":{"detail":{},"id":"1","type":"WEB"},"uuid":"1","version":"0"},"level":"info","msg":"received event","time":"2024-07-03T09:19:59Z"}`,
			want:    []string{"received", "event"},
			pattern: "<_>received event<_>",
		},
		{
			line:    `{"code":200,"message":"OK","data":{"id":"1","name":"foo"}}`,
			want:    []string{"OK"},
			pattern: "<_>OK<_>",
		},
		{
			line:    `{"time":"2024-07-03T10:48:10.58330448Z","level":"INFO","msg":"successfully discovered 15 agent IP addresses","git_commit":"1","git_time":"2024-06-26T06:59:04Z","git_modified":true,"go_os":"linux","go_arch":"arm64","process_generation":"ea2d9b41-0314-4ddc-a415-f8af2d80a32c","hostname_fqdn":"foobar","hostname_short":foobar","private_ips":["10.0.132.23"],"num_vcpus":8,"kafka_enabled":true,"service_protocol":"VIRTUALENV_ZONE_LOCAL","module":"agent_resolver","ip_addresses":[{"Hostname":"10.0.100.152","Port":8080},{"Hostname":"10.0.41.210","Port":8080},{"Hostname":"10.0.212.83","Port":8080},{"Hostname":"10.0.145.77","Port":8080},{"Hostname":"10.0.59.71","Port":8080},{"Hostname":"10.0.224.219","Port":8080},{"Hostname":"10.0.103.29","Port":8080},{"Hostname":"10.0.86.220","Port":8080},{"Hostname":"10.0.154.82","Port":8080},{"Hostname":"10.0.9.213","Port":8080},{"Hostname":"10.0.240.157","Port":8080},{"Hostname":"10.0.166.11","Port":8080},{"Hostname":"10.0.230.22","Port":8080},{"Hostname":"10.0.123.239","Port":8080},{"Hostname":"10.0.233.210","Port":8080}]}`,
			want:    []string{"successfully", "discovered", "15", "agent", "IP", "addresses"},
			pattern: "<_>successfully discovered 15 agent IP addresses<_>",
		},
	}

	tokenizer := newJSONTokenizer(param)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, state := tokenizer.Tokenize(tt.line, nil, nil)
			require.Equal(t, tt.want, got)
			pattern := tokenizer.Join(got, state)
			require.Equal(t, tt.pattern, pattern)
		})
	}
}
