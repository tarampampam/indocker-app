import {Emitter} from 'mitt' // https://github.com/developit/mitt

declare module '@vue/runtime-core' {
  // provide typings for `this.$emitter`
  interface ComponentCustomProperties {
    $emitter: Emitter<any>
  }
}
