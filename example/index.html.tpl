<!DOCTYPE html>
<html>
<head>
  <script type="importmap">{{.importmap}}</script>
</head>
<body>
  <button id="fire">🎉</button>
  <script type="module">
    import confetti from 'confetti';
    document.getElementById('fire').onclick = () => confetti();
  </script>
</body>
</html>
