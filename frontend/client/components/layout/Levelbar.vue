<template>
  <nav class="level">
    <div class="level-left">
      <div class="level-item">
        <h3 class="subtitle is-5">
          <strong>{{ name }}</strong>
        </h3>
      </div>
    </div>

    <div class="level-right is-hidden-mobile">
      <div class="level-item">
        <nav class="breadcrumb is-right">
          <ul>
            <li v-for="(item, index) in list" v-bind:class="{ 'is-active': (index + 1 === list.length) }">
              <router-link :to="item.path">{{item.name}}</router-link>
            </li>
          </ul>
        </nav>
      </div>
    </div>

  </nav>
</template>

<script>
export default {
  data () {
    return {
      list: null
    }
  },

  created () {
    this.getList()
  },

  computed: {
    name () {
      return this.$route.name
    }
  },

  methods: {
    getList () {
      let matched = this.$route.matched.filter(item => item.name)
      let first = matched[0]
      if (first && (first.name !== 'Home' || first.path !== '')) {
        matched = [{ name: 'Home', path: '/' }].concat(matched)
      }
      this.list = matched
    }
  },

  watch: {
    $route () {
      this.getList()
    }
  }
}
</script>
