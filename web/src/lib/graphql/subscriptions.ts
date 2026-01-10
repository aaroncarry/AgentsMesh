import { gql } from "@apollo/client";
import { POD_FRAGMENT, MESSAGE_FRAGMENT, RUNNER_FRAGMENT } from "./queries";

// Pod Subscriptions
export const POD_UPDATED = gql`
  ${POD_FRAGMENT}
  subscription PodUpdated($podKey: String!) {
    podUpdated(podKey: $podKey) {
      ...PodFields
    }
  }
`;

export const POD_OUTPUT = gql`
  subscription PodOutput($podKey: String!) {
    podOutput(podKey: $podKey) {
      podKey
      data
      timestamp
    }
  }
`;

export const POD_STATUS_CHANGED = gql`
  subscription PodStatusChanged($podKey: String!) {
    podStatusChanged(podKey: $podKey) {
      podKey
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

export const POD_JOINED_CHANNEL = gql`
  subscription PodJoinedChannel($channelId: ID!) {
    podJoinedChannel(channelId: $channelId) {
      channelId
      podKey
      agentType
      timestamp
    }
  }
`;

export const POD_LEFT_CHANNEL = gql`
  subscription PodLeftChannel($channelId: ID!) {
    podLeftChannel(channelId: $channelId) {
      channelId
      podKey
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
      currentPods
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
