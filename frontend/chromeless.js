const { Chromeless } = require('chromeless')

async function run() {
  const chromeless = new Chromeless({ waitTimeout: 2000 })

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

  // 4. approve policy change request
  await chromeless
    .click('a[href="#/requests"]')
    .wait('button[class~="delete"]')
    .click('button[class~="delete"]')
    .wait(500)
    .click('button[class~="delete"]')
    .wait(500)
    // this hash should now be valid (after the request creation above)
    .type('dde8892ad346de84', 'input')
    .press(13)
    .wait(500)

  screenshot = await chromeless
    .click('div p button[class*="button is-success"]')
    .wait('input[placeholder="Enter an unseal key"]')
    .type('not_a_real_uneal_key', 'input[placeholder="Enter an unseal key"]')
    .press(13)
    .wait(500)
    .screenshot()
  console.log(screenshot)

  // close notification
  await chromeless.click('button[class~="delete"]').wait(500)

  await chromeless.end()
}

run().catch(console.error.bind(console))
