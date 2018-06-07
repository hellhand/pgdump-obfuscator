package main

type Configuration struct {
	Obfuscations []TargetedObfuscation
}

// TODO: read from file?
var Config *Configuration = &Configuration{
	Obfuscations: []TargetedObfuscation{
		TargetedObfuscation{
			Target{Table: "public.\"user\"", Column: "email_address"},
			ScrambleEmail,
		},
		TargetedObfuscation{
			Target{Table: "public.\"user\"", Column: "first_name"},
			ScrambleBytes,
		},
		TargetedObfuscation{
			Target{Table: "public.\"user\"", Column: "last_name"},
			ScrambleBytes,
		},
	},
}
