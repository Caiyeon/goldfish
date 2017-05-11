<template>
  <footer class="footer">
    <div class="container">
      <div class="content has-text-centered">
        <p class="social">
          <a>
            <iframe src="https://ghbtns.com/github-btn.html?user=caiyeon&repo=goldfish&type=watch&count=true&size=large&v=2" frameborder="0" scrolling="0" width="140px" height="30px"></iframe>
          </a>
          <a>
            <iframe src="https://ghbtns.com/github-btn.html?user=caiyeon&repo=goldfish&type=star&count=true&size=large" frameborder="0" scrolling="0" width="140px" height="30px"></iframe>
          </a>
          <a>
            <iframe src="https://ghbtns.com/github-btn.html?user=caiyeon&repo=goldfish&type=fork&count=true&size=large" frameborder="0" scrolling="0" width="140px" height="30px"></iframe>
          </a>
        </p>
        <p><span class="icon"><i class="fa fa-code"></i></span> with <span class="icon"><i class="fa fa-heart"></i></span> by <a href="https://github.com/caiyeon">caiyeon</a>.</p>
        <p>Code licensed under <a :href="'https://github.com/' + repository + '/blob/master/LICENSE'">{{ license }}</a>.</p>
        <p>Frontend boilerplate made by <a href="https://github.com/fundon">fundon</a>.</p>
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
@import '~bulma/sass/utilities/variables';
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
