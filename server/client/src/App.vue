<template>
  <form
    class="bg-slate-800 p-4 flex justify-center items-center"
    :action="handleSearch"
    @click.prevent="onSubmit"
  >
    <input class="" v-model="search" type="text" />
    <!-- <input type="submit" value="search" @click="handleSearch"> -->
    <button @click="handleSearch">Search</button>
  </form>

  <div class="table w-full ...">
    <div class="table-header-group ...">
      <div class="table-row">
        <div class="table-cell text-left ...">From</div>
        <div class="table-cell text-left ...">To</div>
        <div class="table-cell text-left ...">Subject</div>
      </div>
    </div>
    
    <div class="table-row-group">
      <div
        class="table-row item"
        v-for="{ _source } in data"
        v-bind:key="_source._id"
        @click="getBody(_source.body)"
      >
        <div class="table-cell ...">{{ _source.from }}</div>
        <div class="table-cell ...">{{ _source.to }}</div>
        <div class="table-cell ...">{{ _source.subject }}</div>
        
        
      </div>
    </div>
  </div>
</template>

<script>
import axios from "axios";
export default {
  name: "App",
  data() {
    return {
      URL: "http://localhost:4001",
      search: "",
      data: null,
      body: ""
    };
  },
  async mounted() {
    let response = await axios.get(`${this.URL}/${this.search}`);
    let json = JSON.parse(response.data);
    const {
      hits: { hits },
    } = json;
    this.data = hits;
  },
  methods: {
    async handleSearch() {
      let response = await axios.get(`http://localhost:4001/${this.search}`);
      let json = JSON.parse(response.data);
      const {
        hits: { hits },
      } = json;
      this.data = hits;
      this.body = ""
      console.log(this.URL);
    },
    getBody(body){
      this.body = body;
    }
  },
};
</script>

<style>
li {
  list-style: none;
}
.item{
  cursor: pointer;
}
</style>
