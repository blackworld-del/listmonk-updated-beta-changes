<template>
  <section class="smtp-overview">
    <header class="columns page-header">
      <div class="column is-6">
        <h1 class="title is-4">{{ $t('settings.smtp.overview') }}</h1>
      </div>
      <div class="column has-text-right">
        <b-button type="is-primary" icon-left="download" @click="exportCSV">
          {{ $t('globals.buttons.exportCSV') }}
        </b-button>
      </div>
    </header>

    <b-table :data="overview" :loading="loading" hoverable striped>
      <b-table-column v-slot="props" field="name" :label="$t('settings.mailserver.profileName')" :td-attrs="$utils.tdID">
        <router-link :to="{ name: 'smtpProfileDetail', params: { id: props.row.id } }">
          <strong>{{ props.row.name }}</strong>
        </router-link>
      </b-table-column>

      <b-table-column v-slot="props" field="host" label="SMTP Host">
        {{ props.row.host }}
      </b-table-column>

      <b-table-column v-slot="props" field="enabled" :label="$t('globals.fields.status')">
        <b-tag :type="props.row.enabled ? 'is-success' : 'is-danger'" small>
          {{ props.row.enabled ? $t('globals.buttons.enabled') : $t('globals.buttons.disabled') }}
        </b-tag>
      </b-table-column>

      <b-table-column v-slot="props" field="sentToday" :label="$t('settings.smtp.sentToday')" numeric>
        {{ $utils.formatNumber(props.row.sentToday || 0) }}
      </b-table-column>

      <b-table-column v-slot="props" field="totalSent" :label="$t('settings.smtp.totalSent')" numeric>
        {{ $utils.formatNumber(props.row.totalSent || 0) }}
      </b-table-column>

      <b-table-column v-slot="props" field="totalFailed" :label="$t('settings.smtp.failedEmails')" numeric>
        <span :class="{ 'has-text-danger': props.row.totalFailed > 0 }">{{ $utils.formatNumber(props.row.totalFailed) }}</span>
      </b-table-column>

      <b-table-column v-slot="props" field="successRate" :label="$t('settings.smtp.successRate')" numeric>
        <span :class="{ 'has-text-success': props.row.successRate >= 99, 'has-text-danger': props.row.successRate < 95 }">
          {{ (props.row.successRate || 0).toFixed(1) }}%
        </span>
      </b-table-column>
    </b-table>
  </section>
</template>

<script>
import Vue from 'vue';
import { getSMTPOverview } from '../api';

export default Vue.extend({
  data() {
    return { overview: [], loading: false };
  },

  methods: {
    async loadData() {
      this.loading = true;
      try {
        this.overview = await getSMTPOverview();
      } finally {
        this.loading = false;
      }
    },

    exportCSV() {
      window.location.href = '/api/smtp-profiles/stats/export?format=csv';
    },
  },

  created() {
    this.loadData();
  },
});
</script>
