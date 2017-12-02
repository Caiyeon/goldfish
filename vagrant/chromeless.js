const chromeless = new Chromeless({ remote: true })

const screenshot = await chromeless
  .setViewport({width: 1920, height: 1080, scale: 1})
  .goto('https://vault-ui.io/#/')
  .wait(500)
  .screenshot()

console.log(screenshot)

await chromeless.end()
