package main

import (
	"os"

	app "github.com/Mozlook/fotobudka-backend/internal/app/api"
)

func main() {
	if err := app.Run(); err != nil {
		_, _ = os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}
}
