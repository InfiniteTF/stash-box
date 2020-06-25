/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

import { RoleEnum } from "./globalTypes";

// ====================================================
// GraphQL query operation: Users
// ====================================================

export interface Users_queryUsers_users {
  id: string;
  name: string;
  /**
   * Should not be visible to other users
   */
  email: string | null;
  /**
   * Should not be visible to other users
   */
  roles: RoleEnum[] | null;
  /**
   * Should not be visible to other users
   */
  api_key: string | null;
  /**
   * Calls to the API from this user over a configurable time period
   */
  api_calls: number;
}

export interface Users_queryUsers {
  users: Users_queryUsers_users[];
}

export interface Users {
  queryUsers: Users_queryUsers;
}
