// package auth provides easy to use authentication primitives for
// GondulAPI. While implementing this is trivial, this package provides
// primitives that can be embedded directly into your own structs. It comes
// with two different solutions: Either full embedding, or pre-made
// functions that you can just pass everything to.
//
// The reason there are two options is to allow trivial embedding when your
// data type is a struct, but also use the same code if your data type is
// not (e.g.: if it's an array). The benefit of the latter is limited, but
// ensures all auth checking uses the same code.
//
// This all uses gondulapi.Config.HTTPUser and HTTPPw so user/pass is
// configured in the config file for now. This is likely to be expanded
// upon in the future.
package auth

import (
	"github.com/gathering/gondulapi"
)

// ReadPublic is used to allow GET requests without passwords, but enforce
// (global) auth for all other requests. To use this, simply add
// *auth.ReadPublic to your object struct. (the struct is empty on purpose,
// it only exists for easy embedding)
type ReadPublic struct{}

// Private enforces user/password for both read and write-operations.
type Private struct{}

func CheckReadPublic(basepath string, element string, method string, user string, password string) error {
	if method == "GET" {
		return nil
	}
	if user == gondulapi.Config.HTTPUser && password == gondulapi.Config.HTTPPw {
		return nil
	}
	return gondulapi.Errorf(401, "Auth error")
}


func CheckPrivate(basepath string, element string, method string, user string, password string) error {
	if user == gondulapi.Config.HTTPUser && password == gondulapi.Config.HTTPPw {
		return nil
	}
	return gondulapi.Errorf(401, "Auth error")
}
// Auth implements the gondulapi.Auther interface
func (dummy *ReadPublic) Auth(basepath string, element string, method string, user string, password string) error {
	return CheckReadPublic(basepath, element, method, user, password)
}
// Auth implements the gondulapi.Auther interface
func (dummy *Private) Auth(basepath string, element string, method string, user string, password string) error {
	return CheckPrivate(basepath, element, method, user, password)
}
