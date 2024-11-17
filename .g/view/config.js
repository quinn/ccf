/**
 * @typedef {Object} ConfigObject
 * @property {string} viewFilename
 * @property {string} funcName */

/**
 * @param {Object} options
 * @param {string} options.method
 * @param {string} options.path
 * @returns {ConfigObject} */
function config({ funcName, ...rest }) {
    const viewFilename = convertCase('kebab', funcName)
    return { viewFilename, funcName, ...rest }
}
