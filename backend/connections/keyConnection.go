package connections

// func handleConnectionCredentials(appEnv string, getSecret func(string) (string, error)) (string, error) {
// 	if appEnv == "dev" {
// 		fmt.Println("Starting in dev mode")
// 		jsonFile, err := os.Open("connection.json")
// 		if err != nil {
// 			return "", fmt.Errorf("error opening connection.json: %v", err)
// 		}
// 		defer jsonFile.Close()

// 		var config map[string]string
// 		if err := json.NewDecoder(jsonFile).Decode(&config); err != nil {
// 			return "", fmt.Errorf("error decoding connection.json: %v", err)
// 		}
// 		return config["connectionString"], nil
// 	} else {
// 		connectionString, err := getSecret("ConnectionString")
// 		if err != nil {
// 			return "", fmt.Errorf("error getting secret from key vault: %v", err)
// 		}
// 		return connectionString, nil
// 	}
// }
