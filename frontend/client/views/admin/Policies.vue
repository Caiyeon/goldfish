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
                </tr>
              </thead>
              <tbody>
                <tr v-for="(entry, index) in policies">
                  <td width="34">
                    <span class="icon">
                    <a @click="getPolicyRules(index)">
                      <i class="fa fa-info"></i>
                    </a>
                    </span>
                  </td>
                  <td>
                    {{ entry }}
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
            <textarea class="textarea" placeholder="Select a policy" v-model="policyRules"></textarea>
          </p>
        </article>
      </div>

    </div>

  </div>
</template>

<script>
export default {
  data () {
    return {
      policies: [],
      policyRules: '',
      selectedIndex: -1
    }
  },

  mounted: function () {
    this.$http.get('/api/policies').then((response) => {
      this.policies = response.data.result
    })
    .catch((error) => {
      this.$onError(error)
    })
  },

  methods: {
    getPolicyRules: function (index) {
      this.policyRules = ''
      this.$http.get('/api/policies/' + this.policies[index]).then((response) => {
        this.policyRules = response.data.result
      })
      .catch((error) => {
        this.$onError(error)
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
