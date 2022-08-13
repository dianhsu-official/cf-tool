/**
 * Easy way to get languages list from Codeforces.
 *
 * Usage:
 *  1. Open https://codeforces.com/problemset/customtest
 *  2. Open Browser Developer Tool.
 *  3. Paste this script into Console and Run.
 */

let options = document.querySelectorAll("#pageContent > form > table > tbody > tr > td:nth-child(2) > div:nth-child(1) > select > option");
let res = "";
for(let i = 0; i < options.length; ++i){
    res += `"${options[i].value}": "${options[i].innerText}",\n`;
}
console.log(res);
