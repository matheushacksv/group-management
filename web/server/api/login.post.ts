// ponytail: single admin from env, move users to a table when more people need access
export default defineEventHandler(async (event) => {
  const { email, password } = await readBody<{ email?: string, password?: string }>(event)
  const { adminEmail, adminPasswordHash } = useRuntimeConfig(event)

  if (!adminEmail || !adminPasswordHash) {
    throw createError({ statusCode: 500, statusMessage: 'Admin user not configured' })
  }

  // hash is stored base64 encoded because "$" in .env gets interpolated
  const hash = Buffer.from(adminPasswordHash, 'base64').toString()
  const valid = email === adminEmail && typeof password === 'string' && await verifyPassword(hash, password)
  if (!valid) {
    throw createError({ statusCode: 401, statusMessage: 'Invalid credentials' })
  }

  await setUserSession(event, { user: { email } })
  return { ok: true }
})
