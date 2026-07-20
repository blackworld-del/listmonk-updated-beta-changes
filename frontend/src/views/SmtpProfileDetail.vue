<template>
  <section class="smtp-profile-detail">
    <b-loading :active="loading" />

    <header class="columns page-header">
      <div class="column is-8">
        <router-link :to="{ name: 'smtpProfiles' }" class="is-size-7">
          <b-icon icon="arrow-left" size="is-small" /> {{ $t('settings.smtp.name') }}
        </router-link>
        <h1 class="title is-4" v-if="profile">{{ profile.name }}</h1>
      </div>
      <div class="column has-text-right" v-if="profile">
        <b-tag :type="profile.enabled ? 'is-success' : 'is-danger'">
          {{ profile.enabled ? $t('globals.buttons.enabled') : $t('globals.buttons.disabled') }}
        </b-tag>
      </div>
    </header>

    <div v-if="stats" class="columns is-multiline">
      <div class="column is-12">
        <div class="tile is-ancestor">
          <div class="tile is-vertical is-12">
            <div class="tile">
              <div class="tile is-parent">
                <article class="tile is-child notification">
                  <p class="title is-3">{{ $utils.formatNumber(stats.sentToday || 0) }}</p>
                  <p class="heading">{{ $t('settings.smtp.sentToday') }}</p>
                </article>
              </div>
              <div class="tile is-parent">
                <article class="tile is-child notification">
                  <p class="title is-3">{{ $utils.formatNumber(stats.sentWeek || 0) }}</p>
                  <p class="heading">{{ $t('settings.smtp.sentThisWeek') }}</p>
                </article>
              </div>
              <div class="tile is-parent">
                <article class="tile is-child notification">
                  <p class="title is-3">{{ $utils.formatNumber(stats.sentMonth || 0) }}</p>
                  <p class="heading">{{ $t('settings.smtp.sentThisMonth') }}</p>
                </article>
              </div>
              <div class="tile is-parent">
                <article class="tile is-child notification">
                  <p class="title is-3">{{ $utils.formatNumber(stats.totalSent || 0) }}</p>
                  <p class="heading">{{ $t('settings.smtp.totalSent') }}</p>
                </article>
              </div>
            </div>
            <div class="tile">
              <div class="tile is-parent">
                <article class="tile is-child notification">
                  <p class="title is-3 has-text-danger">{{ $utils.formatNumber(stats.totalFailed || 0) }}</p>
                  <p class="heading">{{ $t('settings.smtp.failedEmails') }}</p>
                </article>
              </div>
              <div class="tile is-parent">
                <article class="tile is-child notification">
                  <p class="title is-3" :class="{ 'has-text-success': stats.successRate >= 99, 'has-text-danger': stats.successRate < 95 }">
                    {{ (stats.successRate || 0).toFixed(1) }}%
                  </p>
                  <p class="heading">{{ $t('settings.smtp.successRate') }}</p>
                </article>
              </div>
              <div class="tile is-parent">
                <article class="tile is-child notification">
                  <p class="title is-5">{{ stats.avgSendTimeSeconds ? (stats.avgSendTimeSeconds).toFixed(2) + 's' : '-' }}</p>
                  <p class="heading">{{ $t('settings.smtp.avgSendTime') }}</p>
                </article>
              </div>
              <div class="tile is-parent">
                <article class="tile is-child notification">
                  <p class="title is-5">{{ stats.lastSentAt ? $utils.niceDate(stats.lastSentAt, true) : '-' }}</p>
                  <p class="heading">{{ $t('settings.smtp.lastEmailSent') }}</p>
                </article>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="column is-6">
        <div class="box">
          <h3 class="title is-6">{{ $t('settings.smtp.generalInfo') }}</h3>
          <table class="table is-fullwidth is-narrow">
            <tr><td class="has-text-grey">{{ $t('settings.mailserver.profileName') }}</td><td>{{ profile.name }}</td></tr>
            <tr><td class="has-text-grey">SMTP Host</td><td>{{ profile.host }}:{{ profile.port }}</td></tr>
            <tr><td class="has-text-grey">{{ $t('settings.mailserver.username') }}</td><td>{{ profile.username }}</td></tr>
            <tr><td class="has-text-grey">{{ $t('settings.mailserver.encryption') }}</td><td>{{ profile.encryption }}</td></tr>
            <tr><td class="has-text-grey">{{ $t('settings.mailserver.fromName') }}</td><td>{{ profile.fromName || '-' }}</td></tr>
            <tr><td class="has-text-grey">{{ $t('campaigns.fromAddress') }}</td><td>{{ profile.fromEmail || '-' }}</td></tr>
            <tr><td class="has-text-grey">Reply-To</td><td>{{ profile.replyTo || '-' }}</td></tr>
            <tr><td class="has-text-grey">{{ $t('globals.fields.status') }}</td>
              <td>
                <b-tag :type="profile.enabled ? 'is-success' : 'is-danger'" small>
                  {{ profile.enabled ? $t('globals.buttons.enabled') : $t('globals.buttons.disabled') }}
                </b-tag>
              </td>
            </tr>
          </table>
        </div>
      </div>

      <div class="column is-6">
        <div class="box">
          <h3 class="title is-6">{{ $t('settings.smtp.dailyStats') }}</h3>
          <div v-if="daily && daily.length">
            <table class="table is-fullwidth is-narrow is-size-7">
              <thead>
                <tr><th>{{ $t('globals.fields.date') }}</th><th class="has-text-right">{{ $t('settings.smtp.sent') }}</th><th class="has-text-right">{{ $t('settings.smtp.failed') }}</th><th class="has-text-right">{{ $t('settings.smtp.successRate') }}</th></tr>
              </thead>
              <tbody>
                <tr v-for="d in daily" :key="d.date">
                  <td>{{ d.date }}</td>
                  <td class="has-text-right">{{ $utils.formatNumber(d.sent) }}</td>
                  <td class="has-text-right"><span :class="{ 'has-text-danger': d.failed > 0 }">{{ d.failed }}</span></td>
                  <td class="has-text-right"><span :class="{ 'has-text-success': d.successRate >= 99 }">{{ d.successRate.toFixed(1) }}%</span></td>
                </tr>
              </tbody>
            </table>
          </div>
          <div v-else class="has-text-grey-light">{{ $t('globals.messages.noData') }}</div>
        </div>
      </div>

      <div class="column is-12">
        <div class="box">
          <h3 class="title is-6">{{ $t('settings.smtp.recentCampaigns') }}</h3>
          <b-table :data="campaigns" :loading="loadingCampaigns" hoverable>
            <b-table-column v-slot="props" field="name" :label="$t('globals.terms.campaign')">
              <router-link :to="{ name: 'campaign', params: { id: props.row.id } }">{{ props.row.name }}</router-link>
            </b-table-column>
            <b-table-column v-slot="props" field="startedAt" :label="$t('campaigns.startedAt')">
              {{ props.row.startedAt ? $utils.niceDate(props.row.startedAt, true) : '-' }}
            </b-table-column>
            <b-table-column v-slot="props" field="endTime" :label="$t('campaigns.ended')">
              {{ props.row.endTime ? $utils.niceDate(props.row.endTime, true) : '-' }}
            </b-table-column>
            <b-table-column v-slot="props" field="sent" :label="$t('campaigns.sent')" numeric>
              {{ $utils.formatNumber(props.row.sent) }}
            </b-table-column>
            <b-table-column v-slot="props" field="failed" :label="$t('settings.smtp.failed')" numeric>
              <span :class="{ 'has-text-danger': props.row.failed > 0 }">{{ props.row.failed }}</span>
            </b-table-column>
            <b-table-column v-slot="props" field="status" :label="$t('globals.fields.status')">
              <b-tag :class="props.row.status" small>{{ $t(`campaigns.status.${props.row.status}`) }}</b-tag>
            </b-table-column>
          </b-table>
        </div>
      </div>

      <div class="column is-12">
        <div class="box">
          <h3 class="title is-6">{{ $t('settings.smtp.recentActivity') }}</h3>
          <b-table :data="activity" :loading="loadingActivity" hoverable>
            <b-table-column v-slot="props" field="createdAt" :label="$t('globals.fields.date')" width="180">
              {{ $utils.niceDate(props.row.createdAt, true) }}
            </b-table-column>
            <b-table-column v-slot="props" field="eventType" :label="$t('settings.smtp.eventType')" width="200">
              <b-tag :type="eventTagType(props.row.eventType)" small>{{ props.row.eventType }}</b-tag>
            </b-table-column>
            <b-table-column v-slot="props" field="message" :label="$t('globals.fields.message')">
              {{ props.row.message }}
            </b-table-column>
          </b-table>
        </div>
      </div>
    </div>

    <div v-if="!profile && !loading" class="has-text-grey-light">{{ $t('globals.messages.notFound', { name: 'SMTP profile' }) }}</div>
  </section>
</template>

<script>
import Vue from 'vue';
import { getSMTPProfile, getSMTPProfileStats } from '../api';

export default Vue.extend({
  data() {
    return {
      profile: null,
      stats: null,
      daily: [],
      campaigns: [],
      activity: [],
      loading: false,
      loadingCampaigns: false,
      loadingActivity: false,
    };
  },

  methods: {
    eventTagType(type) {
      if (!type) return '';
      if (type.includes('success') || type.includes('completed') || type.includes('enabled')) return 'is-success';
      if (type.includes('fail') || type.includes('error') || type.includes('timeout') || type.includes('disabled')) return 'is-danger';
      return 'is-info';
    },

    async fetchData() {
      const id = this.$route.params.id;
      if (!id) return;

      this.loading = true;
      try {
        const [profile, statsData] = await Promise.all([
          getSMTPProfile(id),
          getSMTPProfileStats(id),
        ]);
        this.profile = profile;
        this.stats = statsData.stats;
        this.daily = statsData.daily || [];
        this.campaigns = statsData.campaigns || [];
        this.activity = statsData.activity || [];
      } finally {
        this.loading = false;
      }
    },
  },

  created() {
    this.fetchData();
  },

  watch: {
    '$route.params.id': 'fetchData',
  },
});
</script>
