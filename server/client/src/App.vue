<template>
  <!-- Header -->
  <form class="bg-gray-600 p-4 grid grid-cols-3 grid-rows-1 shadow-sm" @submit.prevent="handleSearch">
  <h1 class="ml-20 font-medium text-white text-xl">Santhz <span class="font-light text-base">mail</span></h1>
    <div class="flex align-middle justify-center col-start-2 col-end-3">
      <input class="bg-gray-200 rounded-full px-3 py-1 text-sm text-gray-700 w-60 " v-model="search" type="text" />
      <input class="ml-2 text-base rounded-sm text-gray-700 cursor-pointer" type="submit"  @click="handleSearch" value="🔍" >
    </div>
  </form>
  <!-- mail -->
  <section class="grid grid-cols-6 px-4 py-2 rounded-sm">
    <div class=" grid grid-cols-3 col-start-2 col-end-6 shadow-sm">
      <div class="bg-gray-100 text-sm font-medium text-gray-500 p-2 grid grid-cols-3 row row-span-1 col-start-1 col-end-4">
        <span >From</span>
        <span >To</span>
        <span >Subject</span>
      </div>
      <EmailView v-for="{_source} in mails" :key="_source._id" :email="_source" />
      
    </div>
  </section>
</template>

<script>
import EmailView from './components/Email.vue'
import axios from "axios";

export default {
  name: "App",
  components:{
    EmailView
  },
  data() {
    return {
      URL: "http://localhost:4001",
      search: "",
      mails: null,
    };
  },
  async mounted() {
    let response = await axios.get(`${this.URL}/${this.search}`);
    let json = JSON.parse(response.data);
    const {
      hits: { hits },
    } = json;
    this.mails = hits;
  },
  methods: {
    async handleSearch() {
      let response = await axios.get(`${this.URL}/${this.search}`);
      let json = JSON.parse(response.data);
      const {
        hits: { hits },
      } = json;
      this.mails = hits;
      this.body = ""
      console.log(this.URL);
    }
  },
};
</script>

<style>
  input{
    outline: none;
    border: none;
  }
</style>
