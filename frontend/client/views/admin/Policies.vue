<template>
  <div>
    <div class="tile is-ancestor">

      <div class="tile is-parent is-vertical is-6">
        <article class="tile is-child box">
          <div class="table-responsive">
            <table class="table is-striped is-narrow">
              <thead>
                <tr>
                  <th></th>
                  <th>Policy Name</th>
                  <th></th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(entry, index) in policies">
                  <td class="is-icon">
                    <a @click="getPolicyRules(index)">
                      <i class="fa fa-info"></i>
                    </a>
                  </td>
                  <td>
                    {{ entry }}
                  </td>
                  <td class="is-icon">
                    <a>
                      <i class="fa fa-trash-o"></i>
                    </a>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </article>
      </div>

      <div class="tile is-parent is-vertical is-6">
        <article class="tile is-child box">
          <h4 class="title is-4">Policy Rules</h4>
          <p class="control">
            <textarea class="textarea" placeholder="Select a policy">{{ policyRules }}</textarea>
          </p>
        </article>
      </div>

    </div>
  </div>
</template>

<script>
  import Tooltip from 'vue-bulma-tooltip'

  export default {
    components: {
      Tooltip
    },

    data () {
      return {
        csrf: '',
        selectedIndex: -1,
        policies: [],
        policyRules: ''
      }
    },

    computed: {
    },

    filters: {
    },

    mounted: function () {
      this.$http.get('/api/policies').then(function (response) {
        this.policies = response.data.result
        this.csrf = response.headers.get('x-csrf-token')
      }, function (err) {
        console.log(err.body.error)
      })
    },

    methods: {
      getPolicyRules: function (index) {
        this.policyRules = ''
        this.$http.get('/api/policies/' + this.policies[index]).then(function (response) {
          this.policyRules = response.data.result
        }, function (err) {
          console.log(err.body.error)
        })
      }
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
