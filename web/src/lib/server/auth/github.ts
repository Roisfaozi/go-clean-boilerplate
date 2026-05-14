export const github = {
  createAuthorizationURL: (state: string, scopes: string[]) => "",
  validateAuthorizationCode: async (code: string) => ({
    accessToken: () => "placeholder-token",
  }),
};
