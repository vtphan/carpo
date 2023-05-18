<template>
  <div>
    <b-card no-body>
      <b-tabs card>
        <b-tab active>
          <template #title>
            <div>Submission</div>
          </template>
          <div style="float:right; position: absolute; top: 2px; left: calc(100% - 165px);">
            <b-dropdown no-caret>
              <template #button-content>
                <b-icon icon="gear-fill" aria-hidden="true"></b-icon> Order By
              </template>
              <b-dropdown-item href="#" @click="setSorting('creation_time')">Creation Time</b-dropdown-item>
              <b-dropdown-item href="#" @click="setSorting('name')">Name</b-dropdown-item>
            </b-dropdown>
          </div>
          <b-card-text>
            <div>
              <div class="items" >
                  <b-card
                    class="item"
                    v-b-modal = "'myModal'"
                    v-bind:img-src="getImagePath()"
                    img-alt="Card image"
                    img-top
                    style="max-width: 14rem;"
                    v-for="items in message.data" :key="items.id"
                    @click="sendInfo(items)">
                      <b-card-text >
                          From: {{ items.student_name }}
                          <br>
                          SubID: {{ items.id }}
                      </b-card-text>
                      <template #footer>
                          <small class="text-muted">Last Active: {{ timeDiff(items.created_at) }} ago </small>
                      </template>
                  </b-card>
              </div>
            </div>

            <b-modal id="myModal" title="Submission Grading" size="lg" ok-only ok-variant="secondary" ok-title="Cancel">
                <codemirror v-model="selectedSub.code" :options="cmOptions" ref="focusThis" />
                <b-row>
                  <b-col cols="6" >
                    <div style="text-align: left">
                      <b-button class="btn-secondary" @click="flagSubmission(selectedSub)">Flag</b-button>
                    </div>
                  </b-col>
                  <b-col cols="6" >
                    <div style="text-align: right">
                      <b-button-group>
                        <b-button class="btn-success" @click="sendGrade(selectedSub, selectedSub.id, 1)">Correct</b-button>
                        <b-button class="btn-danger" @click="sendGrade(selectedSub, selectedSub.id, 2)">Incorrect</b-button>
                        <b-button class="btn-secondary" @click="sendFeedback(selectedSub, selectedSub.id)">Try Again</b-button>
                      </b-button-group>
                    </div>
                  </b-col>
                </b-row>
            </b-modal>
          </b-card-text>
        </b-tab>
        <b-tab >
          <template #title>
            <div> Flagged <a v-if="flagSubs.data">({{ flagSubs.data.length}})</a>
            </div>
          </template>
          <b-card-text>
            <div>
              <div class="items" >
                  <b-card
                    class="item"
                    v-b-modal = "'flagModal'"
                    v-bind:img-src="getImagePath()"
                    img-alt="Card image"
                    img-top
                    style="max-width: 14rem;"
                    v-for="items in flagSubs.data" :key="items.id"
                    @click="sendInfo(items)">
                      <b-card-text >
                          From: {{ items.student_name }}
                          <br>
                          SubID: {{ items.submission_id }}
                      </b-card-text>
                      <template #footer>
                          <small class="text-muted">Last Active: {{ timeDiff(items.created_at) }} ago </small>
                      </template>
                  </b-card>
              </div>
            </div>
            <b-modal id="flagModal" title="Flagged Submission" size="lg" ok-only ok-variant="secondary" ok-title="Unflag" @ok="Unflag(selectedSub)">
                <codemirror v-model="selectedSub.code" :options="cmOptions" ref="focusThis" />
                <b-row>
                  <b-col cols="6" >
                    <div style="text-align: right">
                      <b-button-group>
                        <b-button class="btn-success" @click="sendGrade(selectedSub, selectedSub.submission_id, 1)">Correct</b-button>
                        <b-button class="btn-danger" @click="sendGrade(selectedSub,selectedSub.submission_id, 2)">Incorrect</b-button>
                        <b-button class="btn-secondary" @click="sendFeedback(selectedSub, selectedSub.submission_id)">Try Again</b-button>
                      </b-button-group>
                    </div>
                  </b-col>
                </b-row>

            </b-modal>
          </b-card-text>
        </b-tab>
      </b-tabs>
    </b-card>
  </div>
</template>

<script>
import { codemirror } from 'vue-codemirror'
import 'codemirror/lib/codemirror.css'
// language
import 'codemirror/mode/python/python.js'

// theme css
import 'codemirror/theme/duotone-light.css'

import * as Config from '../config'
import moment from 'moment'

export default {
  name: 'SubmissionList',
  components: {
    codemirror
  },
  data: () => ({
    token: '',
    message: '',
    flagSubs: '',
    sorting: 'creation_time',
    selectedSub: '',
    cmOptions: {
      autoRefresh: true,
      tabSize: 4,
      styleActiveLine: true,
      lineNumbers: true,
      line: true,
      mode: 'application/x-httpd-python',
      lineWrapping: true,
      theme: 'duotone-light'
    }
  }),
  methods: {
    sendInfo (item) {
      console.log('SendInfo:', item)
      this.selectedSub = item
    },
    getImagePath () {
      return require('../assets/code-block-1.png')
    },
    focusMyElement () {
      this.$refs.focusThis.focus()
    },
    timeDiff (dbTimestamp) {
      return moment.duration(moment().diff(moment(dbTimestamp))).humanize()
      // https://stackoverflow.com/questions/18623783/get-the-time-difference-between-two-datetimes
    },
    flagSubmission (submission) {
      const config = {
        headers: { Authorization: 'Bearer ' + this.$route.query.token }
      }
      let postBody = {
        'student_id': submission.student_id,
        'submission_id': submission.id,
        'problem_id': submission.problem_id
        // 'teacher_id': Number(this.$route.query.id)
      }

      this.$http.post(Config.apiUrl + '/submissions/flag', postBody, config)
        .then(() => {
          alert('This submission is now flagged.')
        })
        .catch(function (error) {
          console.log(error)
          // alert(error)
        })
    },
    Unflag (submission) {
      this.$http.delete(Config.apiUrl + '/submissions/flag', {
        headers: { Authorization: 'Bearer ' + this.$route.query.token },
        data: {flag_id: submission.id}
      })
        .then(() => {
          alert('This submission is now unflagged.')
        })
        .catch(function (error) {
          console.log(error)
          // alert(error)
        })
    },
    sendGrade (submission, id, score) {
      const config = {
        headers: { Authorization: 'Bearer ' + this.$route.query.token }
      }

      let postBody = {
        'student_id': submission.student_id,
        'submission_id': id,
        'problem_id': submission.problem_id,
        // 'teacher_id': Number(this.$route.query.id),
        'score': score,
        'code': submission.code
      }

      var status = score === 1 ? 'Correct.' : 'Incorrect.'

      this.$http.post(Config.apiUrl + '/submissions/grade', postBody, config)
        .then(data => {
          alert('This submission is now graded as ' + status)
        })
    },
    sendFeedback (submission, id) {
      const config = {
        headers: { Authorization: 'Bearer ' + this.$route.query.token }
      }

      let postBody = {
        'student_id': submission.student_id,
        'submission_id': id,
        'problem_id': submission.problem_id,
        // 'teacher_id': this.$route.query.id,
        'code': submission.code
      }

      this.$http.post(Config.apiUrl + '/teachers/feedbacks', postBody, config)
        .then(data => {
          alert('Feedback sent to student.')
        })
    },
    getSubmissionList: function () {
      const config = {
        headers: { Authorization: 'Bearer '.concat(this.$route.query.token) }
      }
      this.$http.get(Config.apiUrl + '/teachers/submissions', config, {params: {'sort_by': this.sorting}})
        .then((response) => {
          this.message = response.data
        })
        .catch(function (error) {
          console.log(error)
          // alert(error)
        })
    },
    getFlaggedSubsList: function () {
      const config = {
        headers: { Authorization: 'Bearer '.concat(this.$route.query.token) }
      }
      console.log(config)
      this.$http.get(Config.apiUrl + '/submissions/flag', config)
        .then((response) => {
          this.flagSubs = response.data
        })
        .catch(function (error) {
          console.log(error)
        })
    },
    setSorting (params) {
      this.sorting = params
      this.getSubmissionList()
      this.getFlaggedSubsList()
    }
  },
  created: function () {
    this.getSubmissionList()
    this.getFlaggedSubsList()
    // setInterval(() => this.getSubmissionList(), 10000)
  }
}
</script>

<style>

/* https://stackoverflow.com/questions/59445065/stack-v-cards-within-n-columns */
.items {
  padding: 5px;
  text-align: left;
  background-color: rgb(206, 209, 212);
}

.item {
  background-color: lightgrey;
  display: inline-block;
  width: 100%;
  margin: 10px;
}

.tab-content .active {
    padding: 0px;
}

button {
  margin: 5px
}

/* Make it responsive */
@media only screen and (max-width: 1000px) {
  .items {
    column-count: 6;
  }
}

@media only screen and (max-width: 600px) {
  .items {
    column-count: 6;
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
