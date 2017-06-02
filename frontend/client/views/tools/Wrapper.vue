<template>
  <div>
    <div class="tile is-ancestor">
      <div class="tile is-parent">
        <article class="tile is-child box">
        </article>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  data () {
    return {
      csrf: ''
    }
  },

  mounted: function () {
    this.$notify({
      title: 'Under Construction',
      message: 'This page doesn\'t work.',
      type: 'warning'
    })
  },

  methods: {
  }
}
</script>

<style scoped>
  .button {
    margin: 5px 0 0;
  }

  .control .button {
    margin: inherit;
  }

  .fa-trash-o {
    color: red;
  }

  .fa-info {
    color: lightskyblue;
  }
</style>
