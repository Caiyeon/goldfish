<template>
  <div>
    <div class="tile is-ancestor">
    <div class="tile is-parent">
    <article class="tile is-child box">

      <div v-for="(pair, index) in bulletinPairs" class="tile is-parent is-marginless is-paddingless">

        <div class="tile is-parent is-vertical is-6">
          <article class="message" v-bind:class="pair[0].type || 'is-primary'">
            <div class="message-header">
              <p>{{ pair[0].title }}</p>
            </div>
            <div class="message-body">
              {{ pair[0].message }}
            </div>
          </article>
        </div>

        <div v-if="pair[1]" class="tile is-parent is-vertical is-6">
          <article class="message" v-bind:class="pair[1].type || 'is-primary'">
            <div class="message-header">
              <p>{{ pair[1].title }}</p>
            </div>
            <div class="message-body">
              {{ pair[1].message }}
            </div>
          </article>
        </div>

      </div>

    </article>
    </div>
    </div>
  </div>
</template>

<script>
export default {
  data () {
    return {
      bulletins: []
    }
  },

  computed: {
    session: function () {
      return this.$store.getters.session
    },
    bulletinPairs: function () {
      var pairs = []
      for (var i = 0; i < this.bulletins.length; i += 2) {
        try {
          pairs.push([this.bulletins[i], this.bulletins[i + 1]])
        } catch (err) {
          pairs.push([this.bulletins[i], null])
        }
      }
      return pairs
    }
  },

  mounted: function () {
    this.$http.get('/v1/bulletins', {
      headers: {'X-Vault-Token': this.session ? this.session.token : ''}
    }).then((response) => {
      this.bulletins = response.data.result
    })
    .catch((error) => {
      this.$onError(error)
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
