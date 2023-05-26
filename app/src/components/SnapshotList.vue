<template>
    <div>
      <b-card no-body>
        <b-tabs card>
          <b-tab active>
            <template #title>
              <div v-on:click="getSnapshotList()"> Snapshot <a v-if="message.data">({{ message.data.length}})</a> </div>
            </template>
              <div style="float:right; position: absolute; top: 6px; left: calc(100% - 165px);">
                <b-dropdown no-caret>
                  <template #button-content>
                    <b-icon icon="gear-fill" aria-hidden="true"></b-icon> Order By
                  </template>
                  <b-dropdown-item href="#" @click="setSorting('creation_time')">LastActive At</b-dropdown-item>
                  <b-dropdown-item href="#" @click="setSorting('name')">Name</b-dropdown-item>
                </b-dropdown>
              </div>
            <b-card-text>
              <div >
                <v-row class="five-cols">
                <!-- <div class="items" > -->
                    <b-card
                      class="item"
                      v-b-modal = "'myModal2'"
                      v-bind:img-src="getImagePath()"
                      img-alt="Card image"
                      style="max-width: 14rem;"
                      img-top v-for="items in message.data" :key="items.id"
                      @click="sendInfo(items)">
                        <b-card-text >
                           {{ items.student_name }}
                        </b-card-text>
                        <template #footer>
                          <small>
                            SUBID: {{ items.id }}
                            <br>
                            PID: {{ items.problem_id }}
                            <br>
                            Last Active: {{ timeDiff(items.created_at) }} ago
                          </small>
                        </template>
                      </b-card>
                  <!-- </div> -->
                </v-row>
              </div>
              <b-modal id="myModal2" title="Snapshot View" size="lg" ok-only ok-variant="secondary" ok-title="Send Feedback" @ok="sendFeedback(selectedSub)">
                <codemirror v-model="selectedSub.code" :options="cmOptions" />
                  <div style="float:right; position: absolute; bottom: -55px; right: calc(100% - 85px);">
                    <b-button class="btn-secondary" @click="watchSubmission(selectedSub); $bvModal.hide('myModal2')">Watch</b-button>
                  </div>
              </b-modal>
            </b-card-text>
          </b-tab>
          <b-tab >
            <template #title>
              <div v-on:click="getWatchedSubsList()"> Watched <a v-if="watchSubs.data">({{ watchSubs.data.length}})</a>
              </div>
            </template>
            <b-card-text>
              <div>
                <!-- <div class="items" > -->
                <v-row class="five-cols">
                    <b-card
                      class="item"
                      v-b-modal = "'watchModal'"
                      v-bind:img-src="getImagePath()"
                      img-alt="Card image"
                      img-top
                      style="max-width: 14rem;"
                      v-for="items in watchSubs.data" :key="items.id"
                      @click="sendInfo(items)">
                        <b-card-text >
                            From: {{ items.student_name }}
                        </b-card-text>
                        <template #footer>
                          <small>
                            SUBID: {{ items.submission_id }}
                            <br>
                            PID: {{ items.problem_id }}
                            <br>
                            Last Active: {{ timeDiff(items.created_at) }} ago
                          </small>
                        </template>
                    </b-card>
                <!-- </div> -->
                </v-row>
              </div>
              <b-modal id="watchModal" title="On Watch Snapshot" size="lg" ok-only ok-variant="secondary" ok-title="Unwatch" @ok="unwatchSub(selectedSub)">
                  <codemirror v-model="selectedSub.code" :options="cmOptions" ref="focusThis" />
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
  name: 'SnapshotList',
  components: {
    codemirror
  },
  data: () => ({
    message: '',
    selectedSub: '',
    watchSubs: '',
    sorting: 'creation_time',
    watchedSub: '',
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
    timeDiff (dbTimestamp) {
      return moment.duration(moment().diff(moment(dbTimestamp))).humanize()
      // https://stackoverflow.com/questions/18623783/get-the-time-difference-between-two-datetimes
    },
    toast (msg) {
      this.$bvToast.toast(`${msg}`, {
        title: `Notification`,
        toaster: 'b-toaster-top-center',
        variant: 'secondary',
        solid: true
      })
    },
    watchSubmission (sub) {
      const config = {
        headers: { Authorization: 'Bearer ' + this.$route.query.token }
      }
      let postBody = {
        'student_id': sub.student_id,
        'submission_id': sub.id,
        'problem_id': sub.problem_id
      }

      this.$http.post(Config.apiUrl + '/snapshots/watch', postBody, config)
        .then(() => {
          // alert('Snapshot  with id ' + sub.student_id + ' is on watch list.')
          this.toast('Snapshot of student with id ' + sub.student_id + ' is on watch list.')
        })
        .catch(function (error) {
          console.log(error)
          // alert(error)
        })
    },
    unwatchSub (sub) {
      this.$http.delete(Config.apiUrl + '/snapshots/watch', {
        headers: { Authorization: 'Bearer ' + this.$route.query.token },
        data: {watch_id: sub.id}
      })
        .then(() => {
          // alert('Snapshot  with id ' + sub.student_id + ' is removed from the watch list.')
          this.toast('Snapshot of student with id ' + sub.student_id + ' is removed from the watch list.')
          this.watchSubs.data = this.watchSubs.data.filter(item => item.id !== sub.id)
        })
        .catch(function (error) {
          console.log(error)
          // alert(error)
        })
    },
    sendFeedback (submission) {
      const config = {
        headers: { Authorization: 'Bearer ' + this.$route.query.token }
      }
      let postBody = {
        'student_id': submission.student_id,
        'submission_id': submission.id,
        'problem_id': submission.problem_id,
        'teacher_id': this.$route.query.id,
        'code': submission.code
      }

      this.$http.post(Config.apiUrl + '/teachers/feedbacks', postBody, config)
        .then(data => {
          // alert('Feedback sent to student.')
          this.toast('Feedback is sent to student.')
        })
    },
    getSnapshotList: function () {
      const config = {
        headers: { Authorization: 'Bearer ' + this.$route.query.token }
      }
      this.$http.get(Config.apiUrl + '/teachers/snapshots', config, {
        params: {
          'sort_by': this.sorting
        }
      })
        .then((response) => {
          console.log('Snapshot: ', response)
          this.message = response.data
        })
        .catch((error) => {
          console.log('Error', error)
          this.toast('Unauthorized Access.')
        })
    },
    getWatchedSubsList: function () {
      const config = {
        headers: { Authorization: 'Bearer '.concat(this.$route.query.token) }
      }
      console.log(config)
      this.$http.get(Config.apiUrl + '/snapshots/watch', config)
        .then((response) => {
          this.watchSubs = response.data
          console.log('watched' + JSON.stringify(this.watchSubs))
        })
        .catch((error) => {
          console.log('Error', error)
          this.toast('Unauthorized Access.')
        })
    },
    setSorting (params) {
      this.sorting = params
      this.getSnapshotList()
      this.getWatchedSubsList()
    }
  },
  created: function () {
    this.getSnapshotList()
    this.getWatchedSubsList()
    // setInterval(() => this.getSnapshotList(), 10000)
  }
}
</script>
<style>

.five-cols {
  display: grid;
  grid-template-columns: repeat(8, 1fr);
  background-color: rgb(206, 209, 212);
  padding: 5px;
  text-align: left;
}

br {
  content: "";
  margin: -2em;
  display: block;
  font-size: 24%;
}

/* https://stackoverflow.com/questions/59445065/stack-v-cards-within-n-columns */
.items {
  column-count: 8;
  padding: 5px;
  text-align: left;
  background-color: rgb(206, 209, 212);
}

.item {
  background-color: lightgrey;
  display: inline-block;
  /* width: 100%; */
  margin: 10px;
}

.tab-content .active {
    padding: 0px;
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
