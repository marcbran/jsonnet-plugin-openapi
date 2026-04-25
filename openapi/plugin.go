package openapi

import "github.com/marcbran/jpoet/pkg/jpoet"

func Plugin() *jpoet.Plugin {
	return jpoet.NewPlugin("openapi", nil)
}
