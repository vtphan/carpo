<template>
  <div id="app">
      <nav class="navbar navbar-expand-lg fixed-top navbar-light bg-light">
        <div class="container">
        <div class="collapse navbar-collapse" id="navbarResponsive">
          <ul class="navbar-nav">
            <li class="nav-item">
              <router-link class="nav-link" to="/">Home</router-link>
            </li>
            <li class="nav-item">
              <router-link class="nav-link" to="/problems">Problems</router-link>
            </li>
            <li class="nav-item">
              <router-link class="nav-link" to="/submissions">Submissions</router-link>
            </li>
            <li class="nav-item">
              <router-link class="nav-link" to="/sanpshots">Snapshots</router-link>
            </li>
          </ul>
        </div>
        </div>
      </nav>
      <div class="myContainer">
         <router-view/>
      </div>

  </div>
</template>

<script>

export default {
  name: 'App'
}
</script>

<style>
#app {
  font-family: 'Avenir', Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  text-align: center;
  color: #2c3e50;
  margin-top: 60px;
  margin-left: 30px;
  margin-right: 30px;
}

#myContainer{
  margin-left: 30px;
  margin-right: 30px;
}

@keyframes fadeIn{
  0%{
    opacity: 0;
  }
  50%{
    opacity: 0.5;
  }
  100%{
    opacity: 1;
  }
}
.moveInUp-leave-active{
  animation: moveInUp .3s ease-in;
}
@keyframes moveInUp{
 0%{
  transform: translateY(0);
 }
  100%{
  transform: translateY(-400px);
 }
}
</style>
