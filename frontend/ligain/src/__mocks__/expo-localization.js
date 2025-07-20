module.exports = {
  locale: 'en',
  locales: ['en', 'fr'],
  isRTL: false,
  getLocales: () => [
    {
      languageCode: 'en',
      countryCode: 'US',
      languageTag: 'en-US',
      decimalSeparator: '.',
      groupingSeparator: ',',
    },
  ],
  getCalendars: () => [
    {
      id: 'gregorian',
      calendar: 'gregorian',
      locale: 'en-US',
    },
  ],
  getTimeZone: () => 'America/New_York',
  is24Hour: () => false,
  getCurrencyCode: () => 'USD',
  getDecimalSeparator: () => '.',
  getGroupingSeparator: () => ',',
}; 