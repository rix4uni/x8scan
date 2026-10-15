package banner

import (
	"fmt"
)

// prints the version message
const version = "v0.0.2"

func PrintVersion() {
	fmt.Printf("Current x8scan version %s\n", version)
}

// Prints the Colorful banner
func PrintBanner() {
	banner := `
          ____                          
   _  __ ( __ ) _____ _____ ____ _ ____ 
  | |/_// __  |/ ___// ___// __  // __ \
 _>  < / /_/ /(__  )/ /__ / /_/ // / / /
/_/|_| \____//____/ \___/ \__,_//_/ /_/`
	fmt.Printf("%s\n%45s\n\n", banner, "Current x8scan version "+version)
}
