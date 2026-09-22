import { rocket } from './datastar-rocket.js';

rocket('demo-counter', {
  mode: 'light',
  props: ({ number, string }) => ({
    count: number.step(1).min(0),
    label: string.trim.default('counter'),
  }),
  setup: ({ $$, observeProps, props }) => {
    $$.count = props.count;
    observeProps(() => {
      $$.count = props.count;
    }, 'count');
  },
  render: ({ html, props: { count, label } }) => {
    return html`
      <div class="stack gap-2">
        <button
          type="button"
          data-on:click="$$count += 1"
          data-text="'${label}: ' + $$count"
        ></button>
        <template data-if="$$count !== ${count}">
          <button
            type="button"
            data-on:click="$$count = ${count}"
          >
            reset
          </button>
        </template>
      </div>
    `;
  },
});
