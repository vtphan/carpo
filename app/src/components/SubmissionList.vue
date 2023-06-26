<template>
  <div>
    <b-card no-body>
      <b-tabs card>
        <b-tab active>
          <template #title>
            <div v-on:click="getSubmissionList()">Submission <a v-if="message.data">({{ message.data.length}})</a></div>
          </template>
          <div style="float:right; position: absolute; top: 6px; left: calc(100% - 165px);">
            <b-dropdown no-caret>
              <template #button-content>
                <b-icon icon="gear-fill" aria-hidden="true"></b-icon> Order By
              </template>
              <b-dropdown-item href="#" @click="setSorting('creation_time')">Creation Time</b-dropdown-item>
              <b-dropdown-item href="#" @click="setSorting('name')">Name</b-dropdown-item>
            </b-dropdown>
          </div>
          <div v-if="isLoading">
                <p>LOADING...</p>
          </div>
          <b-card-text v-else>
            <div >
              <v-row class="five-cols">
              <!-- <div class="items" > -->
                  <b-card
                    class="item"
                    :style="{'border-color': setborderColor(items.created_at)}"
                    v-b-modal = "'myModal'"
                    v-for="items in message.data" :key="items.id"
                    @click="sendInfo(items)">
                      <template #header>
                            SUBID: {{ items.id }}
                            <br>
                            PID: {{ items.problem_id }}
                      </template>
                      <b-card-text >
                          {{ items.student_name }}
                      </b-card-text>
                      <template #footer>
                          <small>
                            Last Active: {{ timeDiff(items.created_at) }} ago
                          </small>
                      </template>
                  </b-card>
              <!-- </div> -->
              </v-row>
            </div>

            <b-modal id="myModal" title="Submission Grading" size="xl" ok-only ok-variant="secondary" ok-title="Cancel">
                <codemirror v-model="selectedSub.code" :options="cmOptions" ref="focusThis" />
                <a> Message: {{ selectedSub.message }} </a>
                <b-row>
                  <b-col cols="6" >
                    <div style="text-align: left">
                      <b-button class="btn-secondary" @click="flagSubmission(selectedSub); $bvModal.hide('myModal')">Flag</b-button>
                    </div>
                  </b-col>
                  <b-col cols="6" >
                    <div style="text-align: right">
                      <b-button-group>
                        <b-button class="btn-success" @click="sendGrade(selectedSub, selectedSub.id, 1); $bvModal.hide('myModal') ">Correct</b-button>
                        <b-button class="btn-danger" @click="sendGrade(selectedSub, selectedSub.id, 2); $bvModal.hide('myModal') ">Incorrect</b-button>
                        <!-- <b-button class="btn-secondary" @click="sendFeedback(selectedSub, selectedSub.id); $bvModal.hide('myModal')">Try Again</b-button> -->
                      </b-button-group>
                    </div>
                  </b-col>
                </b-row>
            </b-modal>
          </b-card-text>
        </b-tab>
        <b-tab >
          <template #title>
            <div v-on:click="getFlaggedSubsList()"> Flagged <a v-if="flagSubs.data">({{ flagSubs.data.length}})</a>
            </div>
          </template>
          <div v-if="isLoading">
                <p>LOADING...</p>
          </div>
          <b-card-text v-else>
            <div>
              <!-- <div class="items" > -->
              <v-row class="five-cols">
                  <b-card
                    class="item"
                    v-b-modal = "'flagModal'"
                    :style="{'border-color': setborderColor(items.created_at)}"
                    v-for="items in flagSubs.data" :key="items.id"
                    @click="sendInfo(items)">
                      <template #header >
                        SUBID: {{ items.submission_id }}
                        <br>
                        PID: {{ items.problem_id }}
                      </template>
                      <b-card-text >
                          {{ items.student_name }}
                      </b-card-text>
                      <template #footer>
                          <small>
                            <br>
                            Last Active: {{ timeDiff(items.created_at) }} ago
                          </small>
                      </template>
                  </b-card>
              <!-- </div> -->
              </v-row>
            </div>
            <b-modal id="flagModal" title="Flagged Submission" size="xl" ok-only ok-variant="secondary" ok-title="Cancel" @ok="Unflag(selectedSub); $bvModal.hide('flagModal')">
                <codemirror v-model="selectedSub.code" :options="cmOptions" ref="focusThis" />
                <a> Message: {{ selectedSub.message }} </a>
                <b-row>
                  <b-col cols="6" >
                    <div style="text-align: left">
                      <b-button class="btn-secondary" @click="Unflag(selectedSub); $bvModal.hide('flagModal')">Unflag</b-button>
                    </div>
                  </b-col>
                  <b-col cols="6" >
                    <div style="text-align: right">
                      <b-button-group>
                        <b-button class="btn-success" @click="sendGrade(selectedSub, selectedSub.submission_id, 1); $bvModal.hide('flagModal')">Correct</b-button>
                        <b-button class="btn-danger" @click="sendGrade(selectedSub,selectedSub.submission_id, 2); $bvModal.hide('flagModal')">Incorrect</b-button>
                        <!-- <b-button class="btn-secondary" @click="sendFeedback(selectedSub, selectedSub.submission_id); $bvModal.hide('flagModal')">Try Again</b-button> -->
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
    isLoading: true,
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
    getColor (value) {
      // value from 0 to 1
      var hue = ((1 - value) * 120).toString(10)
      return ['hsl(', hue, ',100%,50%)'].join('')
    },
    setborderColor (dbTimestamp) {
      // https://stackoverflow.com/questions/7128675/from-green-to-red-color-depend-on-percentage
      const limit = 5
      var ago = moment.duration(moment().diff(moment(dbTimestamp))).minutes()
      if (ago > 5) {
        ago = 5
      }
      var value = ago / limit
      return this.getColor(value)
    },
    toast (msg) {
      this.$bvToast.toast(`${msg}`, {
        title: `Notification`,
        toaster: 'b-toaster-top-center',
        variant: 'secondary',
        autoHideDelay: 2000,
        solid: true
      })
    },
    flagSubmission (submission) {
      const config = {
        headers: { Authorization: 'Bearer ' + this.$route.query.token }
      }
      let postBody = {
        'student_id': submission.student_id,
        'submission_id': submission.id,
        'problem_id': submission.problem_id
      }

      this.$http.post(Config.apiUrl + '/submissions/flag', postBody, config)
        .then(() => {
          // alert('This submission is now flagged.')
          this.toast('Submission with id ' + submission.id + ' is flagged.')
          this.message.data = this.message.data.filter(item => item.id !== submission.id)
        })
        .catch(function (error) {
          console.log(error)
        })
    },
    Unflag (submission) {
      this.$http.delete(Config.apiUrl + '/submissions/flag', {
        headers: { Authorization: 'Bearer ' + this.$route.query.token },
        data: {flag_id: submission.id, submission_id: submission.submission_id}
      })
        .then(() => {
          // alert('This submission is now unflagged.')
          this.toast('Submission with id ' + submission.submission_id + ' is unflagged.')
          this.flagSubs.data = this.flagSubs.data.filter(item => item.id !== submission.id)
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
          // alert('This submission is now graded as ' + status)
          this.toast('Submission with id ' + id + ' is graded as ' + status)
          this.message.data = this.message.data.filter(item => item.id !== id)
          this.flagSubs.data = this.flagSubs.data.filter(item => item.id !== submission.id)
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
        'code': submission.code
      }

      this.$http.post(Config.apiUrl + '/teachers/feedbacks', postBody, config)
        .then(data => {
          // alert('Feedback sent to student.')
          this.toast('Feedback for submission with id ' + id + ' is sent to student.')
          this.message.data = this.message.data.filter(item => item.id !== id)
          this.flagSubs.data = this.flagSubs.data.filter(item => item.id !== submission.id)
        })
    },
    getSubmissionList: function () {
      this.isLoading = true
      const config = {
        headers: { Authorization: 'Bearer '.concat(this.$route.query.token) },
        params: {'sort_by': this.sorting}
      }
      this.$http.get(Config.apiUrl + '/teachers/submissions', config)
        .then((response) => {
          this.message = response.data
          this.isLoading = false
        })
        .catch((error) => {
          console.log('Error', error)
          this.toast('Unauthorized Access.')
        })
    },
    getFlaggedSubsList: function () {
      this.isLoading = true
      const config = {
        headers: { Authorization: 'Bearer '.concat(this.$route.query.token) },
        params: {'sort_by': this.sorting}
      }
      console.log(config)
      this.$http.get(Config.apiUrl + '/submissions/flag', config)
        .then((response) => {
          this.flagSubs = response.data
          this.isLoading = false
        })
        .catch((error) => {
          console.log('Error', error)
          this.toast('Unauthorized Access.')
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
  }
}
</script>

<style>

.five-cols {
  display: grid;
  grid-template-columns: repeat(10, 1fr);
  column-gap: 8px;
  background-color: rgb(206, 209, 212);
  padding: 5px;
  /* text-align: left; */
}

.item .card-header {
  padding: 0.25rem 0.25rem;
  font-weight: 300;
  font-size: 12px;
}

.card-body {
  padding: 2px;
  height: 55px;
}

b-card-text {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.card-footer {
  padding: 0.25rem 0.25rem;
}

small, .small {
    font-size: 11px;
    font-weight: 400;
}

br {
  content: "";
  margin: -2em;
  display: block;
  font-size: 24%;
}

.item {
  background-color: rgb(235, 229, 229);
  display: inline-block;
  width: 100%;
  max-width: 14rem;
  margin: 5px;
  font-size: 12px;
  font-weight: 300;
  border-width: 3px;
}

.tab-content .active {
    padding: 0px;
}

button {
  margin: 5px
}

/* Make it responsive
@media only screen and (max-width: 1000px) {
  .items {
    column-count: 6;
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
    column-count: 2;
  }
} */

</style>
