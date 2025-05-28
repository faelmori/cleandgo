package cli

import (
	"math/rand"
	"os"
	"strings"
)

var banners = []string{
	`
     ____    ___                               __      ____    _____
    /\  _ \ /\_ \                             /\ \    /\  _ \ /\  __ \
    \ \ \/\_\//\ \      __     __      ___    \_\ \   \ \ \L\_\ \ \/\ \
     \ \ \/_/_\ \ \   /'__ \ /'__ \  /' _  \  /'_  \   \ \ \L_L\ \ \ \ \
      \ \ \L\ \\_\ \_/\  __//\ \L\.\_/\ \/\ \/\ \L\ \   \ \ \/, \ \ \_\ \
       \ \____//\____\ \____\ \__/.\_\ \_\ \_\ \___,_\   \ \____/\ \_____\
        \/___/ \/____/\/____/\/__/\/_/\/_/\/_/\/__,_ /    \/___/  \/_____/
`,
}

func GetDescriptions(descriptionArg []string, _ bool) map[string]string {
	var description, banner string
	if descriptionArg != nil {
		if strings.Contains(strings.Join(os.Args[0:], ""), "-h") {
			description = descriptionArg[0]
		} else {
			description = descriptionArg[1]
		}
	} else {
		description = ""
	}
	bannerRandLen := len(banners)
	bannerRandIndex := rand.Intn(bannerRandLen)
	banner = banners[bannerRandIndex]
	return map[string]string{"banner": banner, "description": description}
}
