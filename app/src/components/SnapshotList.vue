<template>
    <div>
        <h3> Snapshots </h3>
        <!-- {{message}} -->
        <div >
            <div class="items" >
                <b-card class="item" v-b-modal = "'myModal2'" v-bind:img-src="getImagePath()" img-alt="Card image" img-top v-for="items in message.data" :key="items.id" @click="sendInfo(items)">
                    <b-card-text >
                        Student Name: {{ items.student_name }}
                        {{ items.id }}
                    </b-card-text>
                    <template #footer>
                        <small class="text-muted">Submitted: {{ items.created_at }}</small>
                    </template>
                </b-card>
            </div>
        </div>

        <b-modal id="myModal2" title="Snapshot View">

            <p class="my-4">
                <pre>
                    {{selectedSub.code}}
                </pre>
            </p>
        </b-modal>
    </div>
</template>

<script>
import * as Config from '../config'

export default {
  name: 'SnapshotList',
  data: () => ({
    message: '',
    selectedSub: ''
  }),
  methods: {
    sendInfo (item) {
      console.log('SendInfo:', item)
      this.selectedSub = item
    },
    getImagePath () {
      return require('../assets/code-block-1.png')
    },
    getSubmissionList: function () {
      this.$http.get(Config.apiUrl + '/teachers/submissions', {
        params: {
          'name': 'Instructor-1',
          'id': 1
        }
      })
        .then((response) => {
        //   console.log('Submission: ', response)
          this.message = response.data
        })
        .catch(function (error) {
          console.log(error)
        })
    }
  },
  created: function () {
    this.getSubmissionList()
  }
}
</script>
<style>

/* https://stackoverflow.com/questions/59445065/stack-v-cards-within-n-columns */
.items {
  column-count: 4;
  column-gap: 10px;
  padding: 0 5px;
  /* font-size: 2em; */
  background-color: rgb(32, 84, 143);
}

.item {
  background-color: lightgrey;
  display: inline-block;
  width: 100%;
  margin: 5px 0;
}

/* Make it responsive */
@media only screen and (max-width: 1000px) {
  .items {
    column-count: 4;
  }
}

@media only screen and (max-width: 600px) {
  .items {
    column-count: 3;
  }
}

@media only screen and (max-width: 400px) {
  .items {
    column-count: 2;
  }
}

@media only screen and (max-width: 100px) {
  .items {
    column-count: 1;
  }
}

</style>
