package contextutil

type contextKey string

const RoleKey = contextKey("role")
const LoggerKey = contextKey("logger")
