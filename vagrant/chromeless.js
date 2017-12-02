const chromeless = new Chromeless({ remote: true })

// take a screenshot of the home page
var screenshot = await chromeless
  .setViewport({width: 1920, height: 1080, scale: 1})
  .goto('https://vault-ui.io/#/')
  .wait(500)
  .screenshot()
console.log(screenshot)

// close notification
await chromeless.click('button[class~="delete"]').wait(500)

// take a screenshot of the login page
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

await chromeless.end()
