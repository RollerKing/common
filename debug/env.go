package debug

import "os"

const env = "ENV"

func IsDebug() bool {
	return os.Getenv(env) == "debug"
}
