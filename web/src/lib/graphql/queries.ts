import { gql } from "@apollo/client";

// User Fragments
export const USER_FRAGMENT = gql`
  fragment UserFields on User {
    id
    email
    username
    name
    avatarUrl
    isActive
    lastLoginAt
    createdAt
  }
`;

// Organization Fragments
export const ORGANIZATION_FRAGMENT = gql`
  fragment OrganizationFields on Organization {
    id
    name
    slug
    logoUrl
    subscriptionPlan
    subscriptionStatus
    createdAt
    updatedAt
  }
`;

// Runner Fragments
export const RUNNER_FRAGMENT = gql`
  fragment RunnerFields on Runner {
    id
    nodeId
    description
    status
    lastHeartbeat
    currentPods
    maxConcurrentPods
    runnerVersion
    hostInfo {
      os
      arch
      memory
      cpuCores
      hostname
    }
    createdAt
    updatedAt
  }
`;

// Pod Fragments
export const POD_FRAGMENT = gql`
  fragment PodFields on Pod {
    id
    podKey
    status
    agentStatus
    initialPrompt
    branchName
    worktreePath
    startedAt
    finishedAt
    lastActivity
    createdAt
    runner {
      id
      nodeId
      status
    }
    agentType {
      id
      name
      slug
    }
    repository {
      id
      name
      fullPath
    }
    ticket {
      id
      identifier
      title
    }
    createdBy {
      id
      username
      name
    }
  }
`;

// Channel Fragments
export const CHANNEL_FRAGMENT = gql`
  fragment ChannelFields on Channel {
    id
    name
    description
    document
    isArchived
    createdAt
    updatedAt
    repository {
      id
      name
    }
    ticket {
      id
      identifier
      title
    }
  }
`;

export const MESSAGE_FRAGMENT = gql`
  fragment MessageFields on ChannelMessage {
    id
    content
    messageType
    metadata
    createdAt
    pod {
      podKey
      agentType {
        name
      }
    }
    user {
      id
      username
      name
      avatarUrl
    }
  }
`;

// Ticket Fragments
export const TICKET_FRAGMENT = gql`
  fragment TicketFields on Ticket {
    id
    number
    identifier
    type
    title
    description
    content
    status
    priority
    dueDate
    startedAt
    completedAt
    createdAt
    updatedAt
    reporter {
      id
      username
      name
      avatarUrl
    }
    assignees {
      id
      username
      name
      avatarUrl
    }
    labels {
      id
      name
      color
    }
    repository {
      id
      name
    }
    parentTicket {
      id
      identifier
      title
    }
  }
`;

// User Queries
export const GET_ME = gql`
  ${USER_FRAGMENT}
  ${ORGANIZATION_FRAGMENT}
  query GetMe {
    me {
      ...UserFields
      organizations {
        ...OrganizationFields
        role
      }
    }
  }
`;

// Organization Queries
export const GET_ORGANIZATION = gql`
  ${ORGANIZATION_FRAGMENT}
  query GetOrganization($id: ID, $slug: String) {
    organization(id: $id, slug: $slug) {
      ...OrganizationFields
      memberCount
    }
  }
`;

export const GET_ORGANIZATION_MEMBERS = gql`
  ${USER_FRAGMENT}
  query GetOrganizationMembers($organizationId: ID!) {
    organizationMembers(organizationId: $organizationId) {
      id
      role
      joinedAt
      user {
        ...UserFields
      }
    }
  }
`;

// Runner Queries
export const GET_RUNNERS = gql`
  ${RUNNER_FRAGMENT}
  query GetRunners($status: String) {
    runners(status: $status) {
      ...RunnerFields
    }
  }
`;

export const GET_RUNNER = gql`
  ${RUNNER_FRAGMENT}
  ${POD_FRAGMENT}
  query GetRunner($id: ID!) {
    runner(id: $id) {
      ...RunnerFields
      activePods {
        ...PodFields
      }
    }
  }
`;

export const GET_AVAILABLE_RUNNERS = gql`
  ${RUNNER_FRAGMENT}
  query GetAvailableRunners {
    availableRunners {
      ...RunnerFields
    }
  }
`;

// Pod Queries
export const GET_PODS = gql`
  ${POD_FRAGMENT}
  query GetPods($filter: PodFilter) {
    pods(filter: $filter) {
      pods {
        ...PodFields
      }
      total
    }
  }
`;

export const GET_POD = gql`
  ${POD_FRAGMENT}
  query GetPod($podKey: String!) {
    pod(podKey: $podKey) {
      ...PodFields
    }
  }
`;

// Channel Queries
export const GET_CHANNELS = gql`
  ${CHANNEL_FRAGMENT}
  query GetChannels($filter: ChannelFilter) {
    channels(filter: $filter) {
      channels {
        ...ChannelFields
      }
      total
    }
  }
`;

export const GET_CHANNEL = gql`
  ${CHANNEL_FRAGMENT}
  ${POD_FRAGMENT}
  query GetChannel($id: ID!) {
    channel(id: $id) {
      ...ChannelFields
      pods {
        ...PodFields
      }
    }
  }
`;

export const GET_CHANNEL_MESSAGES = gql`
  ${MESSAGE_FRAGMENT}
  query GetChannelMessages($channelId: ID!, $limit: Int, $offset: Int) {
    channelMessages(channelId: $channelId, limit: $limit, offset: $offset) {
      messages {
        ...MessageFields
      }
      total
    }
  }
`;

// Ticket Queries
export const GET_TICKETS = gql`
  ${TICKET_FRAGMENT}
  query GetTickets($filter: TicketFilter) {
    tickets(filter: $filter) {
      tickets {
        ...TicketFields
      }
      total
    }
  }
`;

export const GET_TICKET = gql`
  ${TICKET_FRAGMENT}
  query GetTicket($identifier: String!) {
    ticket(identifier: $identifier) {
      ...TicketFields
      childTickets {
        id
        identifier
        title
        status
        priority
        type
      }
    }
  }
`;

// Agent Type Queries
export const GET_AGENT_TYPES = gql`
  query GetAgentTypes {
    agentTypes {
      id
      slug
      name
      description
      launchCommand
      defaultArgs
      isBuiltin
      isActive
    }
  }
`;

// Git Provider Queries
export const GET_GIT_PROVIDERS = gql`
  query GetGitProviders {
    gitProviders {
      id
      providerType
      name
      baseUrl
      isDefault
      isActive
      createdAt
    }
  }
`;

export const GET_REPOSITORIES = gql`
  query GetRepositories($gitProviderId: ID) {
    repositories(gitProviderId: $gitProviderId) {
      id
      name
      fullPath
      defaultBranch
      ticketPrefix
      isActive
      gitProvider {
        id
        name
        providerType
      }
    }
  }
`;

// Labels Query
export const GET_LABELS = gql`
  query GetLabels($repositoryId: ID) {
    labels(repositoryId: $repositoryId) {
      id
      name
      color
    }
  }
`;
