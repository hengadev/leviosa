import { computePermissions } from '#lib/security/permissions.ts'
import { ROLES } from '#lib/types/role.ts'

export const load = async ({ locals }) => {
    const user = locals.user

    const role = user?.role ?? ROLES.visitor;
    const permissions = computePermissions(role);

    return {
        user,
        permissions
    }
}

