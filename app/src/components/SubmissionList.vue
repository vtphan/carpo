<template>
 <!-- eslint-disable-next-line vue/max-attributes-per-line -->
    <div>
        <h3> Submissions </h3>
        <!-- {{message}} -->
        <div >
            <div class="items" >
                <b-card
                  class="item"
                  v-b-modal = "'myModal'"
                  v-bind:img-src="getImagePath()"
                  img-alt="Card image"
                  img-top
                  style="max-width: 20rem;"
                  v-for="items in message.data" :key="items.id"
                  @click="sendInfo(items)">
                  <!-- <img :src=getImagePath()> -->
                    <b-card-text >
                        {{ items.student_name }}
                        {{ items.id }}
                    </b-card-text>
                    <!-- <img :src=getImagePath()> -->
                    <template #footer>
                        <small class="text-muted">{{ items.created_at }}</small>
                    </template>
                </b-card>
            </div>
        </div>

        <b-modal id="myModal" title="Submission Grading" ok-only ok-variant="secondary" ok-title="Cancel">
            <code-mirror v-model="selectedSub.code" />
            <div style="text-align: center">
              <b-button-group>
                <b-button class="btn-success" @click="sendGrade(selectedSub,1)">Correct</b-button>
                <b-button class="btn-danger" @click="sendGrade(selectedSub,2)">Incorrect</b-button>
                <b-button class="btn-secondary" @click="sendFeedback(selectedSub)">Try Again</b-button>
              </b-button-group>
            </div>
        </b-modal>
    </div>
</template>

<script>
// import axios from 'axios'
import { ref } from 'vue'
import { python } from '@codemirror/lang-python'
import CodeMirror from 'vue-codemirror6'

import * as Config from '../config'

export default {
  name: 'SubmissionList',
  components: {
    CodeMirror
  },
  data: () => ({
    message: '',
    selectedSub: '',
    lang: ref(python())
  }),
  methods: {
    sendInfo (item) {
      console.log('SendInfo:', item)
      this.selectedSub = item
    },
    getImagePath () {
      return require('../assets/code-block-1.png')
    },
    sendGrade (submission, score) {
      let postBody = {
        'student_id': submission.student_id,
        'submission_id': submission.id,
        'problem_id': submission.problem_id,
        'score': score,
        'code': submission.code // This doesn't have edited code
      }

      var status = score === 1 ? 'Correct.' : 'Incorrect.'

      this.$http.post(Config.apiUrl + '/submissions/grade', postBody)
        .then(data => {
          alert('This submission is now graded as ' + status)
        })
    },
    sendFeedback (submission) {
      let postBody = {
        'student_id': submission.student_id,
        'submission_id': submission.id,
        'problem_id': submission.problem_id,
        'code': submission.code
      }

      this.$http.post(Config.apiUrl + '/teachers/feedbacks', postBody)
        .then(data => {
          alert('Feedback sent to student.')
        })
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

button {
  margin: 5px
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
