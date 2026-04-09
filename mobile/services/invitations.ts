import { AxiosError } from 'axios';
import { api, handleApiError } from './api';
import { ApiResponse } from '../types';

export interface Invitation {
  id: string;
  token: string;
  sender_id: string;
  sender_email: string;
  sender_name: string;
  status: 'pending' | 'accepted' | 'rejected' | 'expired';
  created_at: string;
  expires_at: string;
}

export interface InvitationResponse {
  invitation: Invitation;
  message: string;
}

export async function validateInvitationToken(token: string): Promise<Invitation> {
  try {
    const response = await api.get<ApiResponse<Invitation>>(`/invitations/${token}/validate`);
    return response.data.data;
  } catch (error) {
    const apiError = handleApiError(error as AxiosError);
    throw new Error(apiError.message);
  }
}

export async function acceptInvitation(token: string): Promise<InvitationResponse> {
  try {
    const response = await api.post<ApiResponse<InvitationResponse>>(`/invitations/${token}/accept`);
    return response.data.data;
  } catch (error) {
    const apiError = handleApiError(error as AxiosError);
    throw new Error(apiError.message);
  }
}

export async function rejectInvitation(token: string): Promise<InvitationResponse> {
  try {
    const response = await api.post<ApiResponse<InvitationResponse>>(`/invitations/${token}/reject`);
    return response.data.data;
  } catch (error) {
    const apiError = handleApiError(error as AxiosError);
    throw new Error(apiError.message);
  }
}

export async function getInvitationByToken(token: string): Promise<Invitation> {
  try {
    const response = await api.get<ApiResponse<Invitation>>(`/invitations/${token}`);
    return response.data.data;
  } catch (error) {
    const apiError = handleApiError(error as AxiosError);
    throw new Error(apiError.message);
  }
}
