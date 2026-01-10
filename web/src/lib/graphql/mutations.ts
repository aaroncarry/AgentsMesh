import { gql } from "@apollo/client";
import {
  POD_FRAGMENT,
  CHANNEL_FRAGMENT,
  MESSAGE_FRAGMENT,
  TICKET_FRAGMENT,
  RUNNER_FRAGMENT,
} from "./queries";

// Auth Mutations
export const LOGIN = gql`
  mutation Login($email: String!, $password: String!) {
    login(email: $email, password: $password) {
      token
      user {
        id
        email
        username
        name
        avatarUrl
      }
    }
  }
`;

export const REGISTER = gql`
  mutation Register($input: RegisterInput!) {
    register(input: $input) {
      token
      user {
        id
        email
        username
        name
      }
    }
  }
`;

export const REFRESH_TOKEN = gql`
  mutation RefreshToken {
    refreshToken {
      token
      expiresAt
    }
  }
`;

// Organization Mutations
export const CREATE_ORGANIZATION = gql`
  mutation CreateOrganization($input: CreateOrganizationInput!) {
    createOrganization(input: $input) {
      id
      name
      slug
    }
  }
`;

export const UPDATE_ORGANIZATION = gql`
  mutation UpdateOrganization($id: ID!, $input: UpdateOrganizationInput!) {
    updateOrganization(id: $id, input: $input) {
      id
      name
      slug
      logoUrl
    }
  }
`;

export const INVITE_MEMBER = gql`
  mutation InviteMember($organizationId: ID!, $email: String!, $role: String!) {
    inviteMember(organizationId: $organizationId, email: $email, role: $role) {
      id
      role
      user {
        id
        email
        username
      }
    }
  }
`;

export const REMOVE_MEMBER = gql`
  mutation RemoveMember($organizationId: ID!, $userId: ID!) {
    removeMember(organizationId: $organizationId, userId: $userId)
  }
`;

// Runner Mutations
export const CREATE_REGISTRATION_TOKEN = gql`
  mutation CreateRegistrationToken($input: CreateRegistrationTokenInput!) {
    createRegistrationToken(input: $input) {
      token
      expiresAt
    }
  }
`;

export const DELETE_RUNNER = gql`
  mutation DeleteRunner($id: ID!) {
    deleteRunner(id: $id)
  }
`;

export const UPDATE_RUNNER_STATUS = gql`
  ${RUNNER_FRAGMENT}
  mutation UpdateRunnerStatus($id: ID!, $status: String!) {
    updateRunnerStatus(id: $id, status: $status) {
      ...RunnerFields
    }
  }
`;

// Pod Mutations
export const CREATE_POD = gql`
  ${POD_FRAGMENT}
  mutation CreatePod($input: CreatePodInput!) {
    createPod(input: $input) {
      ...PodFields
    }
  }
`;

export const TERMINATE_POD = gql`
  mutation TerminatePod($podKey: String!) {
    terminatePod(podKey: $podKey) {
      podKey
      status
    }
  }
`;

export const SEND_POD_INPUT = gql`
  mutation SendPodInput($podKey: String!, $data: String!) {
    sendPodInput(podKey: $podKey, data: $data)
  }
`;

export const RESIZE_POD_TERMINAL = gql`
  mutation ResizePodTerminal($podKey: String!, $rows: Int!, $cols: Int!) {
    resizePodTerminal(podKey: $podKey, rows: $rows, cols: $cols)
  }
`;

// Channel Mutations
export const CREATE_CHANNEL = gql`
  ${CHANNEL_FRAGMENT}
  mutation CreateChannel($input: CreateChannelInput!) {
    createChannel(input: $input) {
      ...ChannelFields
    }
  }
`;

export const UPDATE_CHANNEL = gql`
  ${CHANNEL_FRAGMENT}
  mutation UpdateChannel($id: ID!, $input: UpdateChannelInput!) {
    updateChannel(id: $id, input: $input) {
      ...ChannelFields
    }
  }
`;

export const ARCHIVE_CHANNEL = gql`
  mutation ArchiveChannel($id: ID!) {
    archiveChannel(id: $id) {
      id
      isArchived
    }
  }
`;

export const UNARCHIVE_CHANNEL = gql`
  mutation UnarchiveChannel($id: ID!) {
    unarchiveChannel(id: $id) {
      id
      isArchived
    }
  }
`;

export const JOIN_CHANNEL = gql`
  mutation JoinChannel($channelId: ID!, $podKey: String!) {
    joinChannel(channelId: $channelId, podKey: $podKey)
  }
`;

export const LEAVE_CHANNEL = gql`
  mutation LeaveChannel($channelId: ID!, $podKey: String!) {
    leaveChannel(channelId: $channelId, podKey: $podKey)
  }
`;

export const SEND_CHANNEL_MESSAGE = gql`
  ${MESSAGE_FRAGMENT}
  mutation SendChannelMessage($channelId: ID!, $content: String!, $podKey: String) {
    sendChannelMessage(channelId: $channelId, content: $content, podKey: $podKey) {
      ...MessageFields
    }
  }
`;

// Ticket Mutations
export const CREATE_TICKET = gql`
  ${TICKET_FRAGMENT}
  mutation CreateTicket($input: CreateTicketInput!) {
    createTicket(input: $input) {
      ...TicketFields
    }
  }
`;

export const UPDATE_TICKET = gql`
  ${TICKET_FRAGMENT}
  mutation UpdateTicket($identifier: String!, $input: UpdateTicketInput!) {
    updateTicket(identifier: $identifier, input: $input) {
      ...TicketFields
    }
  }
`;

export const DELETE_TICKET = gql`
  mutation DeleteTicket($identifier: String!) {
    deleteTicket(identifier: $identifier)
  }
`;

export const UPDATE_TICKET_STATUS = gql`
  mutation UpdateTicketStatus($identifier: String!, $status: String!) {
    updateTicketStatus(identifier: $identifier, status: $status) {
      identifier
      status
      startedAt
      completedAt
    }
  }
`;

export const ASSIGN_TICKET = gql`
  mutation AssignTicket($identifier: String!, $userIds: [ID!]!) {
    assignTicket(identifier: $identifier, userIds: $userIds) {
      identifier
      assignees {
        id
        username
        name
      }
    }
  }
`;

// Label Mutations
export const CREATE_LABEL = gql`
  mutation CreateLabel($input: CreateLabelInput!) {
    createLabel(input: $input) {
      id
      name
      color
    }
  }
`;

export const DELETE_LABEL = gql`
  mutation DeleteLabel($id: ID!) {
    deleteLabel(id: $id)
  }
`;

// Git Provider Mutations
export const CREATE_GIT_PROVIDER = gql`
  mutation CreateGitProvider($input: CreateGitProviderInput!) {
    createGitProvider(input: $input) {
      id
      providerType
      name
      baseUrl
      isDefault
    }
  }
`;

export const UPDATE_GIT_PROVIDER = gql`
  mutation UpdateGitProvider($id: ID!, $input: UpdateGitProviderInput!) {
    updateGitProvider(id: $id, input: $input) {
      id
      name
      baseUrl
      isDefault
    }
  }
`;

export const DELETE_GIT_PROVIDER = gql`
  mutation DeleteGitProvider($id: ID!) {
    deleteGitProvider(id: $id)
  }
`;

// Repository Mutations
export const CREATE_REPOSITORY = gql`
  mutation CreateRepository($input: CreateRepositoryInput!) {
    createRepository(input: $input) {
      id
      name
      fullPath
      defaultBranch
      ticketPrefix
    }
  }
`;

export const UPDATE_REPOSITORY = gql`
  mutation UpdateRepository($id: ID!, $input: UpdateRepositoryInput!) {
    updateRepository(id: $id, input: $input) {
      id
      name
      defaultBranch
      ticketPrefix
    }
  }
`;

export const DELETE_REPOSITORY = gql`
  mutation DeleteRepository($id: ID!) {
    deleteRepository(id: $id)
  }
`;

// Agent Credential Mutations
export const SET_AGENT_CREDENTIALS = gql`
  mutation SetAgentCredentials($agentTypeId: ID!, $credentials: JSON!) {
    setAgentCredentials(agentTypeId: $agentTypeId, credentials: $credentials)
  }
`;

export const DELETE_AGENT_CREDENTIALS = gql`
  mutation DeleteAgentCredentials($agentTypeId: ID!) {
    deleteAgentCredentials(agentTypeId: $agentTypeId)
  }
`;
