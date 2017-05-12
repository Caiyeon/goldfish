<template>
  <footer class="footer">
    <div class="container">
      <div class="content has-text-centered">
        <p class="social">
          <a>
            <github-button
            href="https://github.com/caiyeon/goldfish"
            data-show-count="true"
            data-size="large"
            data-icon="octicon-star">Stars</github-button>
          </a>
          <a>
            <github-button
            href="https://github.com/caiyeon/goldfish/fork"
            data-show-count="true"
            data-size="large"
            data-icon="octicon-repo-forked">Forks</github-button>
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
import GithubButton from 'github-buttons/dist/vue/GithubButton'

export default {
  components: {
    GithubButton
  },

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
