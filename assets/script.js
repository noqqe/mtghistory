async function readJsonFromUrl(url) {
    let obj = await (await fetch(url)).json();
    return obj;
}

function drawCircle(context, centerX, centerY, radius, fillColor, strokeColor) {
  // Check if the context is valid
  if (!context || typeof context !== "object") {
    throw new Error("Invalid canvas context");
  }
  context.fillStyle = fillColor;
  context.beginPath();
  context.arc(centerX, centerY, radius, 0, 2 * Math.PI, false);
  context.fill();
  context.lineWidth = 1;
  context.strokeStyle = strokeColor;
  context.stroke();
}


var allCards 
(async () => {
  allCards = await readJsonFromUrl('/assets/allcards.json')

  for (const [year, cards] of Object.entries(allCards)) {
    c = 0;
    canvas = document.getElementById(year);
    canvas.width = innerWidth*0.66;
    ctx = canvas.getContext("2d");
    console.log(allCards);

    // adjust canvas heigth depending on how many dots
    // this is a bad way of caluculating the height, but it works for now
    document.getElementById(year).height = cards.length / 5;

    for (let y = 10; y < canvas.height - 5; y = y + 15) {
      for (let x = 10; x < canvas.width - 5; x = x + 15) {
        c = c + 1;
        if (c < cards.length) {
          if (userCards.includes(cards[c])) {
            drawCircle(ctx, x, y, 5, "#50fa7b", "#50fa7b");
          } else {
            drawCircle(ctx, x, y, 4, "grey", "darkgrey");
          }
        } else {
          break;
        }
      }
     }
    }
})()

