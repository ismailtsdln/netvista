package utils

import "fmt"

func GetBanner(version string) string {
	banner := `
  _   _   _____   _____  __     __  ___   ____    _____   _    
 | \ | | | ____| |_   _| \ \   / / |_ _| / ___|  |_   _| / \   
 |  \| | |  _|     | |    \ \ / /   | |  \___ \    | |  / _ \  
 | |\  | | |___    | |     \ V /    | |   ___) |   | | / ___ \ 
 |_| \_| |_____|   |_|      \_/    |___| |____/    |_|/_/   \_\
                                                                   
                           v%s
             Network & Web Host Visual Recon Tool
                  Created by İsmail Taşdelen
`
	return fmt.Sprintf(banner, version)
}
