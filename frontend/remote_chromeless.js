// Remote usage: go to https://chromeless.netlify.com
// Copy this file in place of the existing script and click run!
// Use only when node is unavailable; it is someone's AWS account after-all

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
  .wait(500)
  .click('button[class~="delete"]')
  .wait(500)

// 4. reject and re-create the request to force a first-approval scenario
// bring up the request via hash
await chromeless
  .click('a[href="#/requests"]')
  .wait('button[class~="delete"]')
  .wait(500)
  .click('button[class~="delete"]')
  .wait(500)
  .type('dde8892ad346de84', 'input')
  .press(13)

// delete the request
await chromeless
  .wait('article div p button[class*="button is-warning"]')
  .click('article div p button[class*="button is-warning"]')
  .wait('article div p button[class*="button is-danger"]')
  .click('article div p button[class*="button is-danger"]')
  .wait('button[class~="delete"]')
  .wait(500)
  .click('button[class~="delete"]')
  .wait(500)

// recreate the request
await chromeless
  .click('a[href="#/policies"]')
  .wait('a[class*="tag is-primary"]')
  .click('a[class*="tag is-primary"]')

await chromeless
  .wait('textarea')
  .clearInput('textarea')
  .type('# Propose changes to policies, and ask admins to approve!', 'textarea')
  .click('div p a[class*="button is-primary is-outlined"]')
  .wait('button[class~="delete"]')
  .wait(500)
  .click('button[class~="delete"]')
  .wait(500)

// 5. approve policy change request
await chromeless
  .click('a[href="#/requests"]')
  .wait('button[class~="delete"]')
  .wait(500)
  .click('button[class~="delete"]')
  .wait(500)
  .type('dde8892ad346de84', 'input')
  .press(13)
  .wait(500)

screenshot = await chromeless
  .wait('article div p button[class*="button is-success"]')
  .click('article div p button[class*="button is-success"]')
  .wait('input[placeholder="Enter an unseal key"]')
  .type('not_a_real_unseal_key', 'input[placeholder="Enter an unseal key"]')
  .press(13)
  .wait(500)
  .screenshot()
console.log(screenshot)

// close notification
await chromeless.click('button[class~="delete"]').wait(500)

// -------------------------------------------------------------------------------------
// 6. gather bulletins screenshot
screenshot = await chromeless
  .click('a[href="#/bulletinboard"]')
  .wait('article[class*="message is-success"]')
  .wait(300)
  .screenshot()
console.log(screenshot)

// 7. token creator screenshot
screenshot = await chromeless
  .click('a[href="#/create-token"]')
  .wait(1000)
  .screenshot()
console.log(screenshot)

// 8. tokens screenshot
screenshot = await chromeless
  .click('a[href="#/users"]')
  // load 300 tokens
  .wait('div > a[class="pagination-next"]')
  .wait(500)
  .click('div > a[class="pagination-next"]')
  // wait for the table to load results from server
  .wait('tbody > tr')
  // open up the info panel for a better screenshot
  .click('a[class="pagination-previous"]')
  .wait(500)
  .screenshot()
console.log(screenshot)

await chromeless.end()
