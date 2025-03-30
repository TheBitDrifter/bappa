package coldbrew

import "os"

// isProd indicates if the application is running in production mode
var isProd = os.Getenv("BAPPA_ENV") == "production"

func IsProd() bool {
	return isProd
}
