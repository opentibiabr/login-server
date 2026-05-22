package serviceerrors

func AdminHint(name string) string {
	switch name {
	case "INVALID_CREDENTIALS":
		return "Verify the account email/name, password and accounts.password SHA-1 hash; reset the account password if the record should be valid."
	case "SERVER_NAME_MISMATCH":
		return "Make SERVER_NAME in the login-server .env match serverName in Canary config.lua, then restart the login-server."
	case "SERVER_CONFIG_INVALID":
		return "Check SERVER_PATH and fix Canary config.lua syntax; the login-server must be able to read serverName from it."
	case "SERVER_CONFIG_NOT_FOUND":
		return "Set SERVER_PATH to the Canary root folder containing config.lua, or copy config.lua.dist to config.lua."
	case "DATABASE_UNAVAILABLE":
		return "Check MYSQL_HOST, MYSQL_PORT, MYSQL_USER, MYSQL_PASS and MYSQL_DBNAME in .env, then verify MySQL is running."
	case "ACCOUNT_DATA_UNAVAILABLE":
		return "Check whether the accounts table schema is up to date for Canary, especially id, email, name, password, type, premdays and lastday."
	case "CHARACTER_LIST_LOAD_FAILED":
		return "Check whether the players table schema is up to date for Canary, especially account_id, name, level, sex, vocation, outfit and lastlogin columns."
	case "SESSION_STORAGE_UNAVAILABLE":
		return "Create or migrate the account_sessions table with id, account_id and expires columns, then try logging in again."
	case "SESSION_CREATE_FAILED":
		return "Check account_sessions permissions/schema and MySQL write access for the login-server database user."
	case "LOGIN_SERVICE_UNAVAILABLE":
		return "Check LOGIN_GRPC_PORT/LOGIN_IP and confirm the login-server gRPC process is running."
	case "EVENT_SCHEDULE_UNAVAILABLE":
		return "Check SERVER_PATH/coreDirectory and ensure the event schedule JSON file exists and is valid."
	case "BOOSTED_DATA_UNAVAILABLE":
		return "Check whether boosted_creature and boosted_boss tables exist and contain one raceid row each."
	case "UNSUPPORTED_REQUEST_TYPE":
		return "The client sent a login request type not handled by this login-server version; check client/server compatibility."
	default:
		return "Check the configured login-server log output for the full technical error and verify the .env values."
	}
}
