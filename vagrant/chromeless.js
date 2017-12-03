const chromeless = new Chromeless({ remote: true, waitTimeout: 2000 })

// -------------------------------------------------------------------------------------
// 1. take a screenshot of the home page
var screenshot = await chromeless
  .setViewport({width: 1920, height: 1080, scale: 1})
  .goto('https://vault-ui.io/#/')
  .wait(500)
  .screenshot()
console.log(screenshot)

// close notification
await chromeless.click('button[class~="delete"]').wait(500)

// 2. take a screenshot of the login page
screenshot = await chromeless
  .click('a[href="#/login"]')
  .wait('input[type="password"]')
  .type('goldfish', 'input[type="password"]')
  .press(13)
  .wait(500)
  .screenshot()
console.log(screenshot)

// close notification
await chromeless.click('button[class~="delete"]').wait(500)
await chromeless.click('button[class~="delete"]').wait(500)

// -------------------------------------------------------------------------------------
// 3. construct a policy change request
await chromeless
  .click('a[href="#/policies"]')
  .wait('a[class*="tag is-primary"]')
  .click('a[class*="tag is-primary"]')

await chromeless
  .clearInput('textarea')
  .type('# Propose changes to policies, and ask admins to approve!', 'textarea')
  .click('div p a[class*="button is-primary is-outlined"]')
  .wait('button[class~="delete"]')
  .click('button[class~="delete"]')
  .wait(500)

await chromeless.end()
