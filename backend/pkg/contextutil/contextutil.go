package contextutil

type contextKey string

const RoleKey = contextKey("role")
const UserIDKey = contextKey("userID")
const LoggerKey = contextKey("logger")
