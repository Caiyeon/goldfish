<template>
  <div>
    <div class="tile is-ancestor">

      <div class="tile is-parent">
        <article class="tile is-child box">
          <h1 class="title">{{ path }}</h1>
          <div class="box">
            <div class="table-responsive">
              <table class="table is-striped is-narrow">
                <thead>
                  <tr>
                    <th>Directory</th>
                    <th></th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="entry in tableData">
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
          </div>
        </article>
      </div>

    </div>
  </div>
</template>

<script>
  import Vue from 'vue'
  import Notification from 'vue-bulma-notification'

  const NotificationComponent = Vue.extend(Notification)

  const openNotification = (propsData = {
    title: '',
    message: '',
    type: '',
    direction: '',
    duration: 4500,
    container: '.notifications'
  }) => {
    return new NotificationComponent({
      el: document.createElement('div'),
      propsData
    })
  }

  export default {
    data () {
      return {
        csrf: '',
        path: '/generic/path',
        tableData: [
          'patha',
          'pathb',
          'pathc'
        ]
      }
    },
    mounted: function () {
      openNotification({
        title: 'Under construction',
        message: 'Page is not implemented yet',
        type: 'danger'
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
</style>
