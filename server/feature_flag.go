package main

func GetFeatureFlag(flagName string) (bool, error) {
	row := db.QueryRow("SELECT status FROM feature_flags WHERE flag = ?;", flagName)

	var b bool

	err := row.Scan(&b)

	if err != nil {
		return false, err
	}

	return b, nil
}
