<template>
    <div>
        <h2> Problem Lists </h2>
        <b-button v-b-toggle="'collapse'" class="m-1" v-for="items in message" :key="items.id" @click="selectProblem(items)" >Problem-{{ items.id }}</b-button>

        <b-collapse id="collapse">
          <b-card>
            <pre>
              {{ selectedProb.question }}
            </pre>
          </b-card>
        </b-collapse>

    </div>
</template>

<script>
import * as Config from '../config'

export default {
  name: 'ProblemList',
  data: () => ({
    message: '',
    selectedProb: ''
  }),
  methods: {
    selectProblem (item) {
      console.log(item)
      this.selectedProb = item
    },
    getProblemList: function () {
      this.$http.get(Config.apiUrl + '/problems/list')
        .then((response) => {
          // console.log(response)
          this.message = response.data
        })
        .catch(function (error) {
          console.log(error)
        })
    }
  },
  created: function () {
    this.getProblemList()
  }
}
</script>
<style>

</style>
