<template>
  <div
    @click="toggleModal"
    class="mt-2 mx-1 py-2 px-2 text-gray-600 cursor-pointer break-words items-center text-xs grid grid-cols-3 col-start-1 col-end-4 border-b"
  >
    <span>{{ email.from }}</span>
    <span>{{ email.to }}</span>
    <span>{{ email.subject }}</span>
  </div>
  <ModalEmail v-if="modal" :toggleModal="toggleModal" :email="email" />
</template>
<script>
import ModalEmail from "./ModalEmail.vue";
export default {
  name: "EmailView",
  data() {
    return {
        modal: false,
    };
  },
  components: {
    ModalEmail,
  },
  props: ["email"],
  methods: {
    toggleModal(){
        this.modal = !this.modal
    }
  }
};
</script>
