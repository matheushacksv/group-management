// Proxy to the Go API: checks the login and adds the API key, so the key never reaches the browser
export default defineEventHandler(async (event) => {
  const { user } = await requireUserSession(event)

  const { apiBase, apiKey } = useRuntimeConfig(event)
  const path = getRouterParam(event, 'path')
  const query = getRequestURL(event).search

  return proxyRequest(event, `${apiBase}/api/${path}${query}`, {
    // X-Actor always overwrites whatever the browser sent: the Go API trusts it as "who did it"
    headers: { 'X-API-Key': apiKey, 'X-Actor': user.email }
  })
})
