package vault

import ()

func (auth AuthInfo) GetBulletins() ([]map[string]interface{}, error) {
	c := GetConfig()

	bulletins, err := auth.ListSecret(c.BulletinPath)
	if err != nil {
		return nil, err
	}

	results := make([]map[string]interface{}, len(bulletins))
	for i, bulletin := range bulletins {
		b, ok := bulletin.(string)
		if ok {
			data, err := auth.ReadSecret(c.BulletinPath + b)
			if err != nil {
				return nil, err
			} else {
				results[i] = data
			}
		}
	}

	return results, nil
}
