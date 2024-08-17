

// type DomainName = "fr" | "com"
// type Email = `${string}@${string}.${DomainName}`


// TODO: make a better email validation
function isValidEmail(email: string) bool {
    return false
}

function isValidPassword(email: string) bool {
    return false
}

function validate(email: string, password: string) {
    if (!isValidEmail(email)) {
        // TODO: find which one of these options is the best
        throw new Error("email invalid")
        return fail(400, { invalid: true });
    }
    // TODO: add the password validation
    return true;
}

export { isValidEmail, isValidPassword, validate }
