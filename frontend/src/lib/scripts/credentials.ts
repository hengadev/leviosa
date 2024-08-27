

// type DomainName = "fr" | "com"
// type Email = `${string}@${string}.${DomainName}`


// TODO: make a better email validation
function isValidEmail(email: string): boolean {
    if (email) return true
    return false
}

function isValidPassword(password: string): boolean {
    if (password) return true
    return false
}

function validate(email: string, password: string): boolean {
    if (!isValidEmail(email)) {
        // TODO: find which one of these options is the best
        // throw new Error("email invalid")
        // return fail(400, { invalid: true });
    }
    if (!isValidPassword(password)) {
        // TODO: find which one of these options is the best
        throw new Error("email invalid")
    }
    return true;
}

export { isValidEmail, isValidPassword, validate }
