package vault

import (
	"errors"
)

func (auth *AuthInfo) WrapData(wrapttl string, data map[string]interface{}) (string, error) {
	client, err := auth.Client()
	if err != nil {
		return "", err
	}
	client.SetToken(vaultToken)

	client.SetWrappingLookupFunc(func(operation, path string) string {
		return wrapttl
	})

	resp, err := client.Logical().Write("/sys/wrapping/wrap", data)
	if err != nil {
		return "", err
	}
	return resp.WrapInfo.Token, nil
}

// to do: Find an optimal way to allow unauthenticated users to unwrap data
func (auth *AuthInfo) UnwrapData(wrappingToken string) (map[string]interface{}, error) {
	client, err := auth.Client()
	if err != nil {
		return nil, err
	}
	client.SetToken(vaultToken)

	// make a raw unwrap call. This will use the token as a header
	resp, err := client.Logical().Unwrap(wrappingToken)
	if err != nil {
		return nil, errors.New("Failed to unwrap provided token " + err.Error())
	}
	return resp.Data, nil
}
