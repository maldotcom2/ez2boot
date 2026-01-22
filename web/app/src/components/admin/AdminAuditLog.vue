<template>
  <div class="audit-log-container">
    <div class="audit-log">Audit Log</div>
  </div>
</template>

<script setup>

const emit = defineEmits(['switch-pane'])

</script>

<style scoped>
.audit-log-container {
  background-color: var(--container-modal);
  padding: 12px;
  border-radius: var(--small-radius);
}

</style>