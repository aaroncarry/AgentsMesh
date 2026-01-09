import { gql } from "@apollo/client";
import { SESSION_FRAGMENT, MESSAGE_FRAGMENT, RUNNER_FRAGMENT } from "./queries";

// Session Subscriptions
export const SESSION_UPDATED = gql`
  ${SESSION_FRAGMENT}
  subscription SessionUpdated($sessionKey: String!) {
    sessionUpdated(sessionKey: $sessionKey) {
      ...SessionFields
    }
  }
`;

export const SESSION_OUTPUT = gql`
  subscription SessionOutput($sessionKey: String!) {
    sessionOutput(sessionKey: $sessionKey) {
      sessionKey
      data
      timestamp
    }
  }
`;

export const SESSION_STATUS_CHANGED = gql`
  subscription SessionStatusChanged($sessionKey: String!) {
    sessionStatusChanged(sessionKey: $sessionKey) {
      sessionKey
      status
      agentStatus
      timestamp
    }
  }
`;

// Channel Subscriptions
export const MESSAGE_RECEIVED = gql`
  ${MESSAGE_FRAGMENT}
  subscription MessageReceived($channelId: ID!) {
    messageReceived(channelId: $channelId) {
      ...MessageFields
    }
  }
`;

export const CHANNEL_UPDATED = gql`
  subscription ChannelUpdated($channelId: ID!) {
    channelUpdated(channelId: $channelId) {
      id
      name
      description
      document
      isArchived
    }
  }
`;

export const SESSION_JOINED_CHANNEL = gql`
  subscription SessionJoinedChannel($channelId: ID!) {
    sessionJoinedChannel(channelId: $channelId) {
      channelId
      sessionKey
      agentType
      timestamp
    }
  }
`;

export const SESSION_LEFT_CHANNEL = gql`
  subscription SessionLeftChannel($channelId: ID!) {
    sessionLeftChannel(channelId: $channelId) {
      channelId
      sessionKey
      timestamp
    }
  }
`;

// Runner Subscriptions
export const RUNNER_STATUS_CHANGED = gql`
  ${RUNNER_FRAGMENT}
  subscription RunnerStatusChanged {
    runnerStatusChanged {
      ...RunnerFields
    }
  }
`;

export const RUNNER_HEARTBEAT = gql`
  subscription RunnerHeartbeat($runnerId: ID!) {
    runnerHeartbeat(runnerId: $runnerId) {
      runnerId
      currentSessions
      timestamp
    }
  }
`;

// Ticket Subscriptions
export const TICKET_UPDATED = gql`
  subscription TicketUpdated($identifier: String!) {
    ticketUpdated(identifier: $identifier) {
      identifier
      title
      description
      status
      priority
      assignees {
        id
        username
        name
      }
      updatedAt
    }
  }
`;

// Organization-wide activity subscription
export const ORGANIZATION_ACTIVITY = gql`
  subscription OrganizationActivity {
    organizationActivity {
      type
      resourceType
      resourceId
      actor {
        id
        username
        name
      }
      timestamp
      data
    }
  }
`;
