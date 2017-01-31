# Tabs

[Tabs](http://bulma.io/documentation/components/tabs) component for Vue Bulma.


## Installation

```console
$ npm install vue-bulma-tabs --save
```


## Examples

```vue
<template>
  <tabs animation="slide" :only-fade="false">
    <tab-pane label="Pictures">Pictures Tab</tab-pane>
    <tab-pane label="Music">Music Tab</tab-pane>
    <tab-pane label="Videos" selected>Video Tab</tab-pane>
    <tab-pane label="Documents" disabled>Document Tab</tab-pane>
  </tabs>
</template>

<script>
import { Tabs, TabPane } from 'vue-bulma-tabs'

export default {
  components: {
    Tabs,
    TabPane
  }
}
</script>
```


## Badges

![](https://img.shields.io/badge/license-MIT-blue.svg)
![](https://img.shields.io/badge/status-stable-green.svg)

---

> [fundon.me](https://fundon.me) &nbsp;&middot;&nbsp;
> GitHub [@fundon](https://github.com/fundon) &nbsp;&middot;&nbsp;
> Twitter [@_fundon](https://twitter.com/_fundon)

