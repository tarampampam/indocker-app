<template>
  <n-layout class="container" has-sider>
    <n-layout-sider>
      <containers-list v-if="containers.length" :containers="containers"/>
      <empty-containers-list v-else/>
    </n-layout-sider>
    <n-layout-content content-style="padding: 0 1em;">
      <n-layout v-if="containers.length">
        <n-space justify="space-between">
          <n-thing><!-- anything --></n-thing>
          <content-switcher/>
        </n-space>
        <router-view/>
      </n-layout>
      <n-skeleton v-else text :repeat="2"/>
    </n-layout-content>
  </n-layout>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import ContainersList from './ContainersList.vue'
import EmptyContainersList from './EmptyContainersList.vue'
import ContentSwitcher from './ContentSwitcher.vue'
import { NLayout, NLayoutContent, NLayoutSider, NSkeleton, NSpace, NThing } from 'naive-ui'

export interface Container {
  id: string
  name: string
  tags: string[]
}

export default defineComponent({
  components: {
    NLayout,
    NLayoutSider,
    NLayoutContent,
    NThing,
    NSpace,
    NSkeleton,
    ContainersList,
    EmptyContainersList,
    ContentSwitcher,
  },
  mounted() {
    if (this.containers && this.containers.length) {
      // redirect to first container, if no container is selected
      if (!this.$route.params.id) {
        this.$router.push({name: 'containers.logs', params: {id: this.containers[0].id}})

        return
      }

      // redirect to first container, if selected container does not exist
      if (!this.containers.find((c) => c.id === this.$route.params.id)) {
        this.$router.push({name: 'containers.logs', params: {id: this.containers[0].id}})

        return
      }
    } else {
      // redirect to containers overview, if no containers are available
      this.$router.push({name: 'containers'})
    }
  },
  data(): {
    containers: Container[]
  } {
    return {
      containers: [
        {
          id: 'id-1',
          name: 'app-app-1',
          tags: ['docker-compose'],
        },
        {
          id: 'id-2',
          name: 'app-app-2',
          tags: ['foo', 'bar'],
        },
      ],
    }
  },
})
</script>

<style lang="scss" scoped>
.container {
  height: 100%;

  .n-layout-sider {
    background-color: transparent;
  }
}
</style>
