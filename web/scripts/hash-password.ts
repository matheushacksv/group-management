// Usage: bun scripts/hash-password.ts <password>
// Prints the value for NUXT_ADMIN_PASSWORD_HASH (same scrypt hash nuxt-auth-utils verifies, base64 encoded)
import { Hash } from '@adonisjs/hash'
import { Scrypt } from '@adonisjs/hash/drivers/scrypt'

const password = process.argv[2]
if (!password) {
  console.error('usage: bun scripts/hash-password.ts <password>')
  process.exit(1)
}

const hash = await new Hash(new Scrypt({})).make(password)
console.log(Buffer.from(hash).toString('base64'))
