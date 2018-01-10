<template>
  <footer class="footer">
    <div class="container is-fluid is-marginless">
      <div class="content has-text-centered">
        <p><span class="icon"><i class="fa fa-code"></i></span> with <span class="icon"><i class="fa fa-heart"></i></span> by <a href="https://github.com/caiyeon">caiyeon</a></p>
        <p>
          <a :href="'https://github.com/' + repository">Source code</a>
          licensed under
          <a :href="'https://github.com/' + repository + '/blob/master/LICENSE'">{{ license }}</a>
        </p>
        <p>Frontend boilerplate made by <a href="https://github.com/fundon">fundon</a></p>
      </div>
    </div>
  </footer>
</template>

<script>
export default {
  data () {
    return this.$store.state.pkg
  }
}
</script>

<style lang="scss">
@import '~bulma/sass/utilities/mixins';

.footer {
  margin-left: 180px;

  @include mobile() {
    margin-left: 0;
  }

  .social a {
    border-bottom: none !important;
  }

  .fa.fa-heart {
    color: red;
  }
}
</style>
